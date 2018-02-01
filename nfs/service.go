package nfs

import (
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	v1 "k8s.io/api/core/v1"
	kubernetesclient "k8s.io/client-go/kubernetes"
)

type NfsService struct {
	Mongo      *mongo.Service
	Context    *mongo.Session
	Kubernetes *kubernetes.Service
	clientset  *kubernetesclient.Clientset
	namespace  string
}

func New(m *mongo.Service, k *kubernetes.Service) *NfsService {
	return &NfsService{
		Mongo:      m,
		Context:    m.NewSession(),
		Kubernetes: k,
		namespace:  "default",
	}
}

func (s *NfsService) getClientset() (*kubernetesclient.Clientset, error) {
	if s.clientset != nil {
		return s.clientset, nil
	}
	var err error
	s.clientset, err = s.Kubernetes.CreateClientset()
	return s.clientset, err
}

func (s *NfsService) CreatePV(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	clientset, err := s.getClientset()
	if err != nil {
		return nil, err
	}
	pv, err = clientset.CoreV1().PersistentVolumes().Create(pv)
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (s *NfsService) DeletePV(name string) error {
	clientset, err := s.getClientset()
	if err != nil {
		return err
	}
	err = clientset.CoreV1().PersistentVolumes().Delete(name, nil)
	if err != nil {
		return err
	}
	return nil
}
