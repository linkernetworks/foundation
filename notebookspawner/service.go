package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/notebookspawner/notebook"

	// import global logger
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"gopkg.in/mgo.v2/bson"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PodNamePrefix = "pod-"

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

func New(c config.Config, m *mongo.MongoService, k *kubernetes.Service) *NotebookSpawnerService {
	// FIXME: provide method to free context
	return &NotebookSpawnerService{c, m, m.NewContext(), k, "default"}
}

func (s *NotebookSpawnerService) Sync(notebookID bson.ObjectId, pod v1.Pod) error {
	podStatus := pod.Status

	info := &entity.NotebookProxyInfo{
		IP: podStatus.PodIP,

		// TODO: extract this as the service configuration
		Port: NotebookContainerPort,

		// TODO: pull the pod info to another section
		Phase:     podStatus.Phase,
		Message:   podStatus.Message,
		Reason:    podStatus.Reason,
		StartTime: podStatus.StartTime,
	}

	q := bson.M{"_id": notebookID}
	m := bson.M{"$set": bson.M{"pod": info}}
	return s.Context.C(entity.NotebookCollectionName).Update(q, m)
}

func (s *NotebookSpawnerService) DeployPod(notebook PodDeployment) error {
	return nil
}

func (s *NotebookSpawnerService) Start(nb *entity.Notebook) error {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return err
	}

	workspace := entity.Workspace{}
	err = s.Context.FindOne(entity.WorkspaceCollectionName, bson.M{"_id": nb.WorkspaceID}, &workspace)
	if err != nil {
		return err
	}

	// TODO: load workspace to ensure the workspace exists
	// workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+nb.WorkspaceID.Hex())
	podName := PodNamePrefix + nb.DeploymentID()

	podFactory := NotebookPodFactory{nb}
	podFactory.NewPod(podName, NotebookPodParameters{
		Image:        nb.Image,
		WorkspaceDir: workspace.Directory,
		WorkingDir:   "/batch",
		Bind:         "0.0.0.0",
		Port:         NotebookContainerPort,
		BaseURL:      s.Config.Jupyter.BaseUrl + "/" + nb.Notebook.ID.Hex(),
	})

	// Start pod for notebook in workspace(batch)
	pod := knb.NewPod(podName)

	_, err = clientset.Core().Pods(s.namespace).Create(&pod)
	if err != nil {
		return err
	}

	var signal = make(chan bool, 1)
	go func() {
		o, stop := trackPod(clientset, podName, s.namespace)
	Watch:
		for {
			pod := <-o
			switch phase := pod.Status.Phase; phase {
			case "Pending":
				// updateNotebookProxyInfo(context, knb.Name, pod.Status)
				// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
				for _, c := range pod.Status.ContainerStatuses {
					waitingReason := c.State.Waiting.Reason
					if waitingReason == "ErrImagePull" || waitingReason == "ImagePullBackOff" {
						logger.Errorf("Container is waiting. Reason %s\n", waitingReason)
						break Watch
					}
				}
			case "Running", "Failed", "Succeeded", "Unknown":
				logger.Infof("Notebook %s is %s\n", podName, phase)
				// updateNotebookProxyInfo(context, knb.Name, pod.Status)
				break Watch
			}

		}
		var e struct{}
		signal <- true
		stop <- e
		close(stop)
		close(signal)
		close(o)
	}()
	return nil
}

func (s *NotebookSpawnerService) Stop(nb *entity.Notebook) error {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return err
	}

	podName := PodNamePrefix + nb.DeploymentID()
	err = clientset.Core().Pods(s.namespace).Delete(podName, metav1.NewDeleteOptions(0))
	if err != nil {
		return err
	}
	return nil
}
