package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/internalservice"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/notebookspawner/notebook"

	"gopkg.in/mgo.v2/bson"
	"path/filepath"
)

/*
import v1 "k8s.io/api/core/v1"

type PodProvider interface {
	NewPod() v1.Pod
}
*/

type NotebookSpawnerService struct {
	Config     config.Config
	Mongo      *mongo.MongoService
	Kubernetes *kubernetes.Service
}

func New(c config.Config, m *mongo.MongoService, k *kubernetes.Service) *NotebookSpawnerService {
	return &NotebookSpawnerService{c, m, k}
}

func (s *NotebookSpawnerService) Sync(notebookID bson.ObjectIdHex, pod v1.Pod) error {
	var context = s.Mongo.NewContext()
	defer context.Close()

	podStatus := pod.Status

	// updateNotebookProxyInfo(context, knb.Name, pod.Status)
	info := &entity.NotebookProxyInfo{
		IP: podStatus.PodIP,

		// TODO: extract this as the service configuration
		Port: notebook.NotebookContainerPort,

		Phase:     podStatus.Phase,
		Message:   podStatus.Message,
		Reason:    podStatus.Reason,
		StartTime: podStatus.StartTime,
	}

	q := bson.M{"_id": notebookID}
	m := bson.M{"$set": bson.M{"pod": info}}
	return context.C(entity.NotebookCollectionName).Update(q, m)
}

/*
func updateNotebookProxyInfo(context *mongo.Context, name string, podStatus v1.PodStatus) error {
}
*/

func (s *NotebookSpawnerService) Start(nb *entity.Notebook) error {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return err
	}

	// TODO: load workspace to ensure the workspace exists
	workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+nb.WorkspaceID.Hex())

	// Start pod for notebook in workspace(batch)
	nbs := internalservice.NewNotebookService(clientset, s.Mongo)

	knb := notebook.KubeNotebook{
		Name:      nb.ID.Hex(),
		Workspace: workspace,
		ProxyURL:  s.Config.Jupyter.BaseUrl,
		Image:     nb.Image,
	}
	if _, err := nbs.Start(knb); err != nil {
		return err
	}
	return nil
}

func (s *NotebookSpawnerService) Stop(nb *entity.Notebook) error {
	clientset, err := s.Kubernetes.CreateClientset()
	if err != nil {
		return err
	}

	nbs := internalservice.NewNotebookService(clientset, s.Mongo)
	knb := notebook.KubeNotebook{Name: nb.ID.Hex()}
	return nbs.Stop(knb)
}
