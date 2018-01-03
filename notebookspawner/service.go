package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PodNamePrefix = "notebook-"

// Object as Pod
type PodFactory interface {
	NewPod(podName string) v1.Pod
}

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
	namespace  string
}

func New(c config.Config, m *mongo.MongoService, k *kubernetes.Service, rs *redis.Service) *NotebookSpawnerService {
	// FIXME: provide method to free context
	return &NotebookSpawnerService{c, m, m.NewContext(), k, "default"}
}

func (s *NotebookSpawnerService) Sync(notebookID bson.ObjectId, pod *v1.Pod) error {
	backend := entity.ProxyBackend{
		// TODO: extract this as the service configuration
		IP:        pod.Status.PodIP,
		Port:      NotebookContainerPort,
		Connected: pod.Status.PodIP != "",
	}

	podInfo := entity.PodInfo{
		Phase:     pod.Status.Phase,
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: pod.Status.StartTime,
	}

	q := bson.M{"_id": notebookID}
	m := bson.M{
		"$set": bson.M{
			"backend": backend,
			"pod":     podInfo,
		},
	}
	return s.Context.C(entity.NotebookCollectionName).Update(q, m)
}

func (s *NotebookSpawnerService) DeployPod(notebook PodDeployment) error {
	return nil
}

func (s *NotebookSpawnerService) Start(nb *entity.Notebook) (*podtracker.PodTracker, error) {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return nil, err
	}

	workspace := entity.Workspace{}
	err = s.Context.FindOne(entity.WorkspaceCollectionName, bson.M{"_id": nb.WorkspaceID}, &workspace)
	if err != nil {
		return nil, err
	}

	// TODO: load workspace to ensure the workspace exists
	// workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+nb.WorkspaceID.Hex())
	podName := PodNamePrefix + nb.DeploymentID()

	podFactory := NotebookPodFactory{nb}

	// volumeMounts subPath should not have a root dir. the correct one is like batches/batch-xxx
	workspaceDir := s.Config.FormatWorkspaceBasename(&workspace)

	pod := podFactory.NewPod(podName, NotebookPodParameters{
		Image:        nb.Image,
		WorkspaceDir: workspaceDir,
		WorkingDir:   "/batch",
		Bind:         "0.0.0.0",
		Port:         NotebookContainerPort,
		BaseURL:      nb.Url,
	})
	// Start tracking first
	podTracker := s.startTracking(clientset, podName, nb)

	_, err = clientset.CoreV1().Pods(s.namespace).Get(podName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// Pod not found. Start a pod for notebook in workspace(batch)
		_, err = clientset.Core().Pods(s.namespace).Create(&pod)
		if err != nil {
			podTracker.Stop()
			return nil, err
		}
	} else if err != nil {
		podTracker.Stop()
		return nil, err
	}
	return podTracker, nil
}

func (s *NotebookSpawnerService) Stop(nb *entity.Notebook) (*podtracker.PodTracker, error) {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return nil, err
	}

	podName := PodNamePrefix + nb.DeploymentID()
	// Start tracking first
	podTracker := s.startTracking(clientset, podName, nb)

	err = clientset.Core().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		podTracker.Stop()
		return nil, err
	}
	return podTracker, nil
}
