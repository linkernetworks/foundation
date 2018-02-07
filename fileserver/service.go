package fileserver

import (
	"errors"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kubernetesclient "k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("FileServer is already stopped")

type FileServerPodDeployment interface {
	entity.PodDeployment
	entity.ProxyInfoProvider
}

type FileServerService struct {
	Config     config.Config
	Mongo      *mongo.Service
	Context    *mongo.Session
	Kubernetes *kubernetes.Service
	Redis      *redis.Service

	clientset *kubernetesclient.Clientset
	namespace string
}

func New(c config.Config, m *mongo.Service, k *kubernetes.Service, rds *redis.Service) *FileServerService {
	return &FileServerService{
		Config:     c,
		Mongo:      m,
		Context:    m.NewSession(),
		Kubernetes: k,
		Redis:      rds,
		namespace:  "default",
	}
}

func (s *FileServerService) getClientset() (*kubernetesclient.Clientset, error) {
	if s.clientset != nil {
		return s.clientset, nil
	}
	var err error
	s.clientset, err = s.Kubernetes.CreateClientset()
	return s.clientset, err
}

func (s *FileServerService) GetPod(fs *entity.FileServer) (*v1.Pod, error) {
	clientset, err := s.getClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(s.namespace).Get(fs.DeploymentID(), metav1.GetOptions{})
}

func (s *FileServerService) Start(fs *entity.FileServer) (tracker *podtracker.PodTracker, err error) {
	workspace := entity.Workspace{}
	err = s.Context.FindOne(entity.WorkspaceCollectionName, bson.M{"_id": fs.WorkspaceID}, &workspace)
	if err != nil {
		return nil, err
	}

	// TODO: load workspace to ensure the workspace exists
	// workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+fs.WorkspaceID.Hex())
	podName := fs.DeploymentID()

	podFactory := FileServerPodFactory{fs}

	pod := podFactory.NewPod(podName, FileServerPodParameters{
		Image: fs.Image,
		Port:  FileServerContainerPort,
	})

	// Start tracking first
	_, err = s.GetPod(fs)
	if kerrors.IsNotFound(err) {
		// Pod not found. Start a pod for fileserver in workspace(batch)
		tracker, err = s.startTracking(podName, fs)
		if err != nil {
			return nil, err
		}

		_, err = s.clientset.CoreV1().Pods(s.namespace).Create(&pod)
		if err != nil {
			tracker.Stop()
			return nil, err
		}
		return tracker, nil

	} else if err != nil {
		// unknown error
		return nil, err
	}

	tracker, err = s.startTracking(podName, fs)
	return tracker, err
}

// Stop returns nil if it's already stopped
func (s *FileServerService) Stop(fs *entity.FileServer) (*podtracker.PodTracker, error) {
	// if it's not created
	_, err := s.GetPod(fs)
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	podName := fs.DeploymentID()

	clientset, err := s.getClientset()
	if err != nil {
		return nil, err
	}

	// force sending a terminating state to document
	q := bson.M{"_id": fs.ID}
	m := bson.M{
		"$set": bson.M{
			"backend.connected": false,
			"pod.phase":         "Terminating",
		},
	}
	s.Context.C(entity.FileServerCollectionName).Update(q, m)

	// We found the pod, let's start a tracker first, and then delete the pod
	podTracker, err := s.startTracking(podName, fs)
	if err != nil {
		return nil, err
	}

	err = clientset.CoreV1().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		podTracker.Stop()
		return nil, ErrAlreadyStopped
	} else if err != nil {
		podTracker.Stop()
		return nil, err
	}
	return podTracker, nil
}

func (s *FileServerService) SyncFromPod(fs *entity.FileServer, pod *v1.Pod) (err error) {
	backend, err := podproxy.NewProxyBackendFromPodStatus(pod, "fileserver")
	if err != nil {
		return err
	}
	podInfo := entity.NewPodInfo(pod)

	q := bson.M{"_id": fs.ID}
	m := bson.M{
		"$set": bson.M{
			"backend": backend,
			"pod":     podInfo,
		},
	}

	err = s.Context.C(entity.FileServerCollectionName).Update(q, m)

	go func() {
		topic := fs.Topic()
		s.Redis.PublishAndSetJSON(topic, fs.NewUpdateEvent(bson.M{
			"backend.connected": pod.Status.PodIP != "",
			"pod.phase":         pod.Status.Phase,
			"pod.message":       pod.Status.Message,
			"pod.reason":        pod.Status.Reason,
			"pod.startTime":     pod.Status.StartTime,
		}))
	}()
	return err
}
