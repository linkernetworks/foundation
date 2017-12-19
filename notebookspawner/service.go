package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/internalservice"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/notebookspawner/notebook"

	v1 "k8s.io/api/core/v1"

	"path/filepath"
)

type PodProvider interface {
	NewPod(podName string) v1.Pod
}

type NotebookSpawnerService struct {
	Config     config.Config
	Mongo      *mongo.MongoService
	Kubernetes *kubernetes.Service
}

func New(c config.Config, m *mongo.MongoService, k *kubernetes.Service) *NotebookSpawnerService {
	return &NotebookSpawnerService{c, m, k}
}

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
