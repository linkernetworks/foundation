package notebookspawner

import (
	"errors"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kubernetesclient "k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("Already Stopped")

type PodLabelProvider interface {
	PodLabels() map[string]string
}

type ProxyInfoProvider interface {
	Host() string
	Port() string
	BaseURL() string
}

type DeploymentIDProvider interface {
	DeploymentID() string
}

type PodDeployment interface {
	DeploymentIDProvider
	PodFactory
}

type NotebookPodDeployment interface {
	PodDeployment
	ProxyInfoProvider
}

type NotebookSpawnerService struct {
	Config     config.Config
	Mongo      *mongo.MongoService
	Context    *mongo.Context
	Kubernetes *kubernetes.Service
	Redis      *redis.Service

	clientset *kubernetesclient.Clientset
	namespace string
}

func New(c config.Config, m *mongo.MongoService, k *kubernetes.Service, rds *redis.Service) *NotebookSpawnerService {
	clientset, err := k.CreateClientset()
	if err != nil {
		panic(err)
	}
	return &NotebookSpawnerService{
		Config:     c,
		Mongo:      m,
		Context:    m.NewContext(),
		Kubernetes: k,
		Redis:      rds,
		namespace:  "default",
		clientset:  clientset,
	}
}

func NewPodInfo(pod *v1.Pod) *entity.PodInfo {
	return &entity.PodInfo{
		Phase:     pod.Status.Phase,
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: pod.Status.StartTime,
	}
}

func (s *NotebookSpawnerService) Sync(notebookID bson.ObjectId, pod *v1.Pod) error {
	backend, err := podproxy.NewProxyBackendFromPodStatus(pod, "notebook")
	if err != nil {
		return err
	}
	podInfo := NewPodInfo(pod)

	q := bson.M{"_id": notebookID}
	m := bson.M{
		"$set": bson.M{
			"backend": backend,
			"pod":     podInfo,
		},
	}
	return s.Context.C(entity.NotebookCollectionName).Update(q, m)
}

func (s *NotebookSpawnerService) Start(nb *entity.Notebook) (tracker *podtracker.PodTracker, err error) {
	workspace := entity.Workspace{}
	err = s.Context.FindOne(entity.WorkspaceCollectionName, bson.M{"_id": nb.WorkspaceID}, &workspace)
	if err != nil {
		return nil, err
	}

	// TODO: load workspace to ensure the workspace exists
	// workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+nb.WorkspaceID.Hex())
	podName := nb.DeploymentID()

	podFactory := NotebookPodFactory{nb}

	// volumeMounts subPath should not have a root dir. the correct one is like batches/batch-xxx
	workspaceDir := s.Config.FormatWorkspaceBasename(&workspace)

	pod := podFactory.NewPod(podName, NotebookPodParameters{
		Image:        nb.Image,
		WorkspaceDir: workspaceDir,
		WorkingDir:   s.Config.Jupyter.WorkingDir,
		Bind:         s.Config.Jupyter.Bind,
		Port:         NotebookContainerPort,
		BaseURL:      nb.Url,
	})
	// Start tracking first
	_, err = s.GetPod(nb)
	if kerrors.IsNotFound(err) {
		// Pod not found. Start a pod for notebook in workspace(batch)
		tracker = s.startTracking(podName, nb)
		_, err = s.clientset.Core().Pods(s.namespace).Create(&pod)
		if err != nil {
			tracker.Stop()
			return nil, err
		}
		return tracker, nil

	} else if err != nil {
		// unknown error
		return nil, err
	}

	tracker = s.startTracking(podName, nb)
	return tracker, nil
}

func (s *NotebookSpawnerService) GetPod(nb *entity.Notebook) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(nb.DeploymentID(), metav1.GetOptions{})
}

// Stop returns nil if it's already stopped
func (s *NotebookSpawnerService) Stop(nb *entity.Notebook) (*podtracker.PodTracker, error) {
	// if it's not created
	_, err := s.GetPod(nb)
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	podName := nb.DeploymentID()

	// We found the pod, let's start a tracker first, and then delete the pod
	podTracker := s.startTracking(podName, nb)
	err = s.clientset.Core().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		podTracker.Stop()
		return nil, ErrAlreadyStopped
	} else if err != nil {
		podTracker.Stop()
		return nil, err
	}
	return podTracker, nil
}
