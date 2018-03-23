package appspawner

import (
	"errors"
	"fmt"

	// "bitbucket.org/linkernetworks/aurora/src/aurora/provision/path"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment/podfactory"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	// "bitbucket.org/linkernetworks/aurora/src/types/container"

	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"bitbucket.org/linkernetworks/aurora/src/workspace"

	"k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("Notebook is already stopped")

type AppSpawner struct {
	Config config.Config

	Factories map[string]entity.WorkspaceAppPodFactory

	Updater *podproxy.ProxyAddressUpdater

	clientset *kubernetes.Clientset
	namespace string
}

func New(c config.Config, clientset *kubernetes.Clientset, rds *redis.Service) *AppSpawner {
	return &AppSpawner{
		Factories: map[string]entity.WorkspaceAppPodFactory{
			"notebook": &podfactory.NotebookPodFactory{
				Config: c.Jupyter,
			},
		},
		Config:    c,
		namespace: "default",
		clientset: clientset,
		Updater: &podproxy.ProxyAddressUpdater{
			Clientset: clientset,
			Namespace: "default",
			Redis:     rds,
			PortName:  "notebook",
		},
	}
}

func (s *AppSpawner) NewPod(app *entity.WorkspaceApp) (*v1.Pod, error) {
	factory, ok := s.Factories[app.ContainerApp.Type]
	if !ok {
		return nil, fmt.Errorf("pod factory for type '%s' is not defined.", app.ContainerApp.Type)
	}
	pod := factory.NewPod(app)

	// attach the primary volumes to the pod spec
	if err := workspace.AttachVolumesToPod(app.Workspace, pod); err != nil {
		return pod, err
	}

	return pod, nil
}

func (s *AppSpawner) Start(ws *entity.Workspace, app *entity.ContainerApp) (tracker *podtracker.PodTracker, err error) {
	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: ws}

	pod, err := s.NewPod(wsApp)

	if err != nil {
		return nil, err
	}

	// Start tracking first
	_, err = s.getPod(wsApp.PodName())
	if kerrors.IsNotFound(err) {
		// Pod not found. Start a pod for notebook in workspace(batch)
		tracker, err = s.Updater.TrackAndSyncUpdate(wsApp)
		if err != nil {
			return nil, err
		}

		_, err = s.clientset.CoreV1().Pods(s.namespace).Create(pod)
		if err != nil {
			tracker.Stop()
			return nil, err
		}
		return tracker, nil

	} else if err != nil {
		// unknown error
		return nil, err
	}

	return s.Updater.TrackAndSyncUpdate(wsApp)
}

func (s *AppSpawner) getPod(name string) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(name, metav1.GetOptions{})
}

// Stop returns nil if it's already stopped
func (s *AppSpawner) Stop(ws *entity.Workspace, app *entity.ContainerApp) (*podtracker.PodTracker, error) {
	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: ws}

	// if it's not created
	_, err := s.getPod(wsApp.PodName())
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	s.Updater.Reset(wsApp)

	// We found the pod, let's start a tracker first, and then delete the pod
	tracker, err := s.Updater.TrackAndSyncDelete(wsApp)
	if err != nil {
		return nil, err
	}

	podName := wsApp.PodName()
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
