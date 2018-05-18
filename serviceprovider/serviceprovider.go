package serviceprovider

import (
	"time"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/config/serviceconfig"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"bitbucket.org/linkernetworks/aurora/src/service/appspawner"
	"bitbucket.org/linkernetworks/aurora/src/service/gearman"
	"bitbucket.org/linkernetworks/aurora/src/service/googlemap"
	hdfsservice "bitbucket.org/linkernetworks/aurora/src/service/hdfs"
	"bitbucket.org/linkernetworks/aurora/src/service/influxdb"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/kudis"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"bitbucket.org/linkernetworks/aurora/src/service/session"
	"bitbucket.org/linkernetworks/aurora/src/service/socketio"
	"bitbucket.org/linkernetworks/aurora/src/service/timer"
	"bitbucket.org/linkernetworks/aurora/src/service/websocket"
	"bitbucket.org/linkernetworks/aurora/src/service/workspacefsspawner"
	"github.com/colinmarc/hdfs"
)

type Container struct {
	Config    config.Config
	Redis     *redis.Service
	Socketio  *socketio.Service
	Timer     *timer.TimerService
	Mongo     *mongo.Service
	Gearman   *gearman.Service
	WebSocket *websocket.WebSocketService
	Kudis     *kudis.Service
	Influxdb  *influxdb.InfluxdbService
	GoogleMap *googlemap.GoogleMapService
	HDFS      *hdfs.Client

	// Kubernetes service for notebooks
	Kubernetes *kubernetes.Service

	AppSpawner *appspawner.AppSpawner

	// The Spawner for fileserver, each workspace has its own fileserver to mount the volume
	WorkspaceSpawner *workspacefsspawner.WorkspaceFileServerSpawner
}

type ServiceDiscoverResponse struct {
	Container map[string]Service `json:"services"`
}

type Service interface{}

type SmartTrackerService struct {
	Redis     serviceconfig.ServiceConfig `json:"redis"`
	Gearman   serviceconfig.ServiceConfig `json:"gearman"`
	Memcached serviceconfig.ServiceConfig `json:"memcached"`
}

func NewRedisService(cf *redis.RedisConfig) *redis.Service {
	logger.Infof("Connecting to redis: %s", cf.Addr())
	return redis.New(cf)
}

func NewInfluxdbService(cf *influxdb.InfluxdbConfig) *influxdb.InfluxdbService {
	logger.Infof("Connecting to influxdb: %s", cf.Url)
	return &influxdb.InfluxdbService{Url: cf.Url}
}

func New(cf config.Config) *Container {
	// setup logger configuration
	logger.Setup(cf.Logger)

	redisService := NewRedisService(cf.Redis)
	socketService := socketio.New(cf.Socketio)

	gcService := timer.New(30 * time.Minute)
	gcService.Bind("cv-client", func() error {
		return redisService.RemoveExpiredClients()
	})
	gcService.Bind("client-subscription", func() error {
		return socketService.CleanUp()
	})
	gcService.Run()

	err := newSessionService(cf.App.Session, cf.Redis.Addr())
	for err != nil {
		logger.Errorf("session initialization failed, retrying: %v", err)
		err = newSessionService(cf.App.Session, cf.Redis.Addr())
		time.Sleep(time.Second * 1)
	}

	logger.Infof("Using jobserver via gPRC: %s", cf.JobServer.Addr())

	logger.Infof("Using kudis via gRPC: %s", cf.Kudis.Addr())

	logger.Infof("Connecting to mongodb: %s", cf.Mongo.Url)
	mongo := mongo.New(cf.Mongo.Url)

	sp := &Container{
		Config:    cf,
		Redis:     redisService,
		Socketio:  socketService,
		Timer:     gcService,
		Mongo:     mongo,
		WebSocket: websocket.NewWebSocketService(),
		Influxdb:  NewInfluxdbService(cf.Influxdb),
		GoogleMap: googlemap.New(cf.GoogleMap.Key),
	}

	if cf.Gearman != nil {
		sp.Gearman = gearman.NewFromConfig(cf.Gearman)
	} else {
		logger.Warnln("Gearman service is not loaded: gearman config is not defined.")
	}

	// services that depends on Kubernetes
	if cf.Kubernetes == nil {
		logger.Warnln("kubernetes service is not loaded: kubernetes config is not defined.")
	} else {
		if cf.Kudis == nil {
			logger.Warnln("kudis service is not loaded: kudis config is not defined.")
		} else {
			sp.Kudis = kudis.New(cf.Kudis)
		}

		sp.Kubernetes = kubernetes.NewFromConfig(cf.Kubernetes)
		clientset, err := sp.Kubernetes.NewClientset()
		if err == nil {
			if cf.Jupyter == nil {
				logger.Warnln("jupyter noteook spawner service is not loaded: jupyter notebook config is not defined.")
			} else {
				sp.AppSpawner = appspawner.New(cf, clientset, sp.Redis, sp.Mongo)
			}
			sp.WorkspaceSpawner = workspacefsspawner.New(cf, sp.Mongo, clientset, sp.Redis)
		} else {
			logger.Errorf("failed to create clientset, worksapce file server/notebook is not enabled: error=%v", err)
		}
	}

	// optional services
	if cf.Hdfs != nil && cf.Hdfs.Host != "" && cf.Hdfs.Enabled {
		sp.HDFS = NewHdfsService(cf)
	}
	return sp
}

func NewHdfsService(cf config.Config) *hdfs.Client {
	url := cf.Hdfs.Addr()
	logger.Info("Connecting to HDFS: ", url, cf.Hdfs.Username)
	client, err := hdfsservice.NewClientForUser(url, cf.Hdfs.Username)

	if err != nil {
		logger.Errorln("Can't connect to HDFS: ", url, cf.Hdfs.Username)
	}

	return client

}

func NewContainer(configPath string) *Container {
	cf := config.MustRead(configPath)
	return New(cf)
}

func (s *Container) DiscoverServices() map[string]Service {
	return map[string]Service{
		"smartTrackerService": SmartTrackerService{
			Redis:     s.Config.Redis.GetPublic(),
			Gearman:   s.Config.Gearman.GetPublic(),
			Memcached: s.Config.Memcached.GetPublic(),
		},
	}
}

func newSessionService(sc *config.SessionConfig, redisAddr string) error {
	return session.NewService(
		sc.Size,
		sc.Protocal,
		redisAddr,
		sc.Password,
		sc.Age,
		sc.KeyPair,
	)
}
