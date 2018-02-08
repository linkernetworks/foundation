package fileserver

import (
	_ "bitbucket.org/linkernetworks/aurora/src/aurora"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"errors"

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
	Session    *mongo.Session
	Kubernetes *kubernetes.Service
	Redis      *redis.Service

	clientset *kubernetesclient.Clientset
	namespace string
}

func New(c config.Config, m *mongo.Service, k *kubernetes.Service, rds *redis.Service) *FileServerService {
	return &FileServerService{
		Config:     c,
		Mongo:      m,
		Session:    m.NewSession(),
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
	ws.PodName = WorkspacePodNamePrefix + ws.ID.Hex()
	_, err := s.GetPod(ws.PodName)
	if kerrors.IsNotFound(err) {
		//Create pod
		volumes := []container.Volume{
			{
				ClaimName: ws.MainVolume.Name,
				Volume: container.VolumeMount{
					Name:      ws.MainVolume.Name,
					MountPath: "/workspace",
				},
			},
		}

		fsParameter := FileServerPodParameters{
			//FIXME for testing, use develop
			//		Image:   FileServerImage + ":" + aurora.ImageTag,
			Image:   FileServerImage + ":develop",
			Port:    FileServerContainerPort,
			Labels:  ws.MainVolume.Labels,
			Volumes: volumes,
		}

		podFactory := FileServerPodFactory{}
		pod := podFactory.NewPod(ws.PodName, fsParameter)
		_, err = s.clientset.CoreV1().Pods(s.namespace).Create(&pod)

		if err != nil {
			return err
		}
		//Update DB (podName)
		s.Session.UpdateById(entity.WorkspaceCollectionName, ws.ID, ws)

	} else if err != nil {
		return err
	}

	return nil
}

func (s *FileServerService) Delete(ws *entity.Workspace) error {
	_, err := s.GetPod(ws.PodName)
	if err != nil {
		return err
		//Create pod
	}
	err = s.clientset.CoreV1().Pods(s.namespace).Delete(ws.PodName, &metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	ws.PodName = ""
	s.Session.UpdateById(entity.WorkspaceCollectionName, ws.ID, ws)
	return nil
}
