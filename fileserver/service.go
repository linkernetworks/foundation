package fileserver

import (
	"errors"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"

	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kubernetesclient "k8s.io/client-go/kubernetes"

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

func (s *FileServerService) GetPod(podName string) (*v1.Pod, error) {
	clientset, err := s.getClientset()
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Pods(s.namespace).Get(podName, metav1.GetOptions{})
}

func (s *FileServerService) WakeUp(ws *entity.Workspace) error {
	_, err := s.GetPod(ws.PodName)
	if kerrors.IsNotFound(err) {
		//Create pod
		//Update DB
	} else if err != nil {
		return err
	}

	return nil
}
