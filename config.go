package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"

	"bitbucket.org/linkernetworks/aurora/src/config/serviceconfig"
	"bitbucket.org/linkernetworks/aurora/src/service/gearman"
	"bitbucket.org/linkernetworks/aurora/src/service/influxdb"
	"bitbucket.org/linkernetworks/aurora/src/utils/netutils"
	"github.com/linkernetworks/logger"
	"github.com/linkernetworks/mongo"
	"github.com/linkernetworks/redis"
)

type Config struct {
	Redis      *redis.RedisConfig       `json:"redis"`
	Gearman    *gearman.GearmanConfig   `json:"gearman"`
	Memcached  *MemcachedConfig         `json:"memcached"`
	Mongo      *mongo.MongoConfig       `json:"mongo"`
	Hdfs       *HdfsConfig              `json:"hdfs"`
	Logger     logger.LoggerConfig      `json:"logger"`
	App        *AppConfig               `json:"app"`
	Jupyter    *JupyterConfig           `json:"jupyter"`
	JobServer  *JobServerConfig         `json:"jobserver"`
	JobUpdater *JobUpdaterConfig        `json:"jobupdater"`
	Migration  *MigrationConfig         `json:"migration"`
	Kudis      *KudisConfig             `json:"kudis"`
	Influxdb   *influxdb.InfluxdbConfig `json:"influxdb"`
	GoogleMap  *GoogleMapConfig         `json:"googlemap"`
	Data       *DataConfig              `json:"data"`
	Features   *FeatureConfig           `json:"features"`

	Socketio *SocketioConfig `json:"socketio"`

	// the version settings of the current application
	Version string `json:"version"`

	// config for kubernetes service, container application instances like
	// jupyter notebook will be created via this service.
	Kubernetes *KubernetesConfig `json:"kubernetes"`
}

// GetWorkspaceRootDir - Get batch process directory
func (c *Config) GetWorkspaceRootDir() string {
	return c.Data.WorkspaceDir
}

//GetArchiveDir - Get archive directory.
func (c *Config) GetArchiveDir() string {
	return c.Data.ArchiveDir
}

//GetImageDir - Get image directory.
func (c *Config) GetImageDir() string {
	return c.Data.ImageDir
}

//GetThumbnailDir - Get thumbnail directory.
func (c *Config) GetThumbnailDir() string {
	return c.Data.ThumbnailDir
}

//GetModelDir - Get model directory.
func (c *Config) GetModelDir() string {
	return c.Data.ModelDir
}

//GetModelArchiveDir - Get model directory.
func (c *Config) GetModelArchiveDir() string {
	return c.Data.ModelArchiveDir
}

func SetupServiceAddressFromInterface(c serviceconfig.ServiceConfig) {
	if reflect.ValueOf(c).IsNil() {
		return
	}
	if c.Unresolved() {
		var inf = c.GetInterface()
		if inf != "" {
			var ip net.IP = netutils.MustFindInterfaceGlobalUnicastIp(inf)
			if ip != nil {
				name := reflect.TypeOf(c).String()
				log.Printf("%s: discovered IP %s on Interface %s\n", name, ip, inf)
				c.SetHost(ip.String())
			}
		}
	}
	c.LoadDefaults()

	if pc := c.GetPublic(); pc != nil {
		SetupServiceAddressFromInterface(pc)
	}
}

// SetupAddressFromInterface will scan the network interface setting if "interface" field
// is defined.
func SetupAddressFromInterface(c *Config) {
	SetupServiceAddressFromInterface(c.Redis)
	SetupServiceAddressFromInterface(c.Gearman)
	SetupServiceAddressFromInterface(c.Memcached)
}

func CanLoadDefaults(c interface{}) bool {
	rf := reflect.ValueOf(c)
	if rf.Kind() != reflect.Ptr {
		return false
	}
	if rf.IsNil() {
		return false
	}
	inf := rf.Interface()
	if inf == nil {
		return false
	}
	_, ok := inf.(serviceconfig.DefaultLoader)
	return ok
}

// LoadDefaults iterates the config fields and calls the load default if it
// implements the interface.
func LoadDefaults(c interface{}) {
	rp := reflect.ValueOf(c)
	if rp.Kind() == reflect.Struct {
		panic(fmt.Errorf("You can not pass value to LoadDefaults. It needs a pointer"))
	}
	if CanLoadDefaults(rp.Interface()) {
		if inf := rp.Interface(); inf != nil {
			if loader, ok := inf.(serviceconfig.DefaultLoader); ok {
				if err := loader.LoadDefaults(); err != nil {
					panic(fmt.Errorf("failed to load default values: %v", err))
				}
				return
			}
		}
	}

	var rv = reflect.ValueOf(c).Elem()
	for i := 0; i < rv.NumField(); i++ {
		rf := rv.Field(i)
		if CanLoadDefaults(rf.Interface()) {
			var inf = rf.Interface()
			if inf == nil {
				continue
			}
			if loader, ok := inf.(serviceconfig.DefaultLoader); ok {
				if err := loader.LoadDefaults(); err != nil {
					panic(fmt.Errorf("failed to load default values: %v", err))
				}
			}
		}
	}
}

func Read(path string) (c Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return c, fmt.Errorf("Failed to open the config file: %v\n", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&c); err != nil {
		return c, fmt.Errorf("Failed to load the config file: %v\n", err)
	}
	SetupAddressFromInterface(&c)

	LoadDefaults(&c)

	return c, nil
}

func MustRead(path string) Config {
	c, err := Read(path)
	if err != nil {
		panic(err)
	}
	return c
}
