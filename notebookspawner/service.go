package notebookspawner

import (
	"errors"

	// "bitbucket.org/linkernetworks/aurora/src/aurora/provision/path"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
	// "bitbucket.org/linkernetworks/aurora/src/types/container"

	kvolume "bitbucket.org/linkernetworks/aurora/src/kubernetes/volume"

	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("Notebook is already stopped")

// Attach the workspace volumes
func AttachWorkspaceVolumes(pod v1.Pod, workspace *entity.Workspace) v1.Pod {
	var volumes = []v1.Volume{}
	var mounts = []v1.VolumeMount{}

	if workspace.PrimaryVolume != nil {
		volumes = append(volumes, kvolume.NewVolume(*workspace.PrimaryVolume))
		mounts = append(mounts, kvolume.NewVolumeMount(*workspace.PrimaryVolume))
	}

	var secondaryVolumes = kvolume.NewVolumes(workspace.SecondaryVolumes)
	volumes = append(volumes, secondaryVolumes...)

	secondaryMounts := kvolume.NewVolumeMounts(workspace.SecondaryVolumes)
	mounts = append(mounts, secondaryMounts...)

	pod.Spec.Volumes = append(pod.Spec.Volumes, volumes...)
	for _, container := range pod.Spec.Containers {
		container.VolumeMounts = append(container.VolumeMounts, mounts...)
	}

	return pod
}

type NotebookSpawnerService struct {
	Config  config.Config
	Session *mongo.Session

	Updater *podproxy.DocumentProxyInfoUpdater

	clientset *kubernetes.Clientset
	namespace string
}

func New(c config.Config, session *mongo.Session, clientset *kubernetes.Clientset, rds *redis.Service) *NotebookSpawnerService {
	return &NotebookSpawnerService{
		Config:    c,
		Session:   session,
		namespace: "default",
		clientset: clientset,
		Updater: &podproxy.DocumentProxyInfoUpdater{
			Clientset:      clientset,
			Namespace:      "default",
			Redis:          rds,
			Session:        session,
			CollectionName: entity.NotebookCollectionName,
			PortName:       "notebook",
		},
	}
}

func (s *NotebookSpawnerService) Start(nb *entity.Notebook) (tracker *podtracker.PodTracker, err error) {
	workspace := entity.Workspace{}
	err = s.Session.FindOne(entity.WorkspaceCollectionName, bson.M{"_id": nb.WorkspaceID}, &workspace)
	if err != nil {
		return nil, err
	}

	// TODO: load workspace to ensure the workspace exists
	// workspace := filepath.Join(s.Config.Data.BatchDir, "batch-"+nb.WorkspaceID.Hex())
	podName := nb.DeploymentID()

	factory := NewNotebookPodFactory(nb, NotebookPodParameters{
		Image:   nb.Image,
		WorkDir: s.Config.Jupyter.WorkingDir,
		Bind:    s.Config.Jupyter.Address,
		Port:    DefaultNotebookContainerPort,
		BaseURL: nb.Url,
		Volumes: workspace.SecondaryVolumes,
	})

	pod := factory.NewPod(podName, map[string]string{
		"service":   "notebook",
		"workspace": nb.WorkspaceID.Hex(),
		"user":      nb.CreatedBy.Hex(),
	})

	pod = AttachWorkspaceVolumes(pod, &workspace)

	// Start tracking first
	_, err = s.getPod(nb)
	if kerrors.IsNotFound(err) {
		// Pod not found. Start a pod for notebook in workspace(batch)
		tracker, err = s.Updater.TrackAndSync(nb)
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

	tracker, err = s.Updater.TrackAndSync(nb)
	return tracker, err
}

func (s *NotebookSpawnerService) getPod(doc types.DeploymentIDProvider) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

// Stop returns nil if it's already stopped
func (s *NotebookSpawnerService) Stop(notebook *entity.Notebook) (*podtracker.PodTracker, error) {
	// if it's not created
	_, err := s.getPod(notebook)
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	s.Updater.Reset(notebook)

	// We found the pod, let's start a tracker first, and then delete the pod
	tracker, err := s.Updater.TrackAndSync(notebook)
	if err != nil {
		return nil, err
	}

	podName := notebook.DeploymentID()
	err = s.clientset.CoreV1().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		defer tracker.Stop()
		if kerrors.IsNotFound(err) {
			return nil, ErrAlreadyStopped
		}
		return nil, err
	}
	return tracker, nil
}
