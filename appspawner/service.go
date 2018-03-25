package appspawner

import (
	"errors"
	"fmt"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment/podfactory"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"bitbucket.org/linkernetworks/aurora/src/workspace"

	"k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("Application is already stopped")

type WorkspaceAppPodFactory interface {
	NewPod(app *entity.WorkspaceApp) *v1.Pod
}

type AppSpawner struct {
	Config config.Config

	Factories map[string]WorkspaceAppPodFactory

	AddressUpdater *ProxyAddressUpdater

	mongo *mongo.Service

	clientset *kubernetes.Clientset
	namespace string
}

func New(c config.Config, clientset *kubernetes.Clientset, rds *redis.Service, m *mongo.Service) *AppSpawner {
	return &AppSpawner{
		Factories: map[string]WorkspaceAppPodFactory{
			"notebook": &podfactory.NotebookPodFactory{Config: c.Jupyter},
		},
		Config:    c,
		namespace: "default",
		clientset: clientset,
		mongo:     m,
		AddressUpdater: &ProxyAddressUpdater{
			Clientset: clientset,
			Namespace: "default",
			Cache:     podproxy.NewDefaultProxyCache(rds),
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
		tracker, err = s.AddressUpdater.TrackAndSyncUpdate(wsApp)
		if err != nil {
			return nil, err
		}

		_, err = s.clientset.CoreV1().Pods(s.namespace).Create(pod)
		if err != nil {
			tracker.Stop()
			return nil, err
		}

		var session = s.mongo.NewSession()
		defer session.Close()
		if err := workspace.AddInstances(session, ws.ID, wsApp.PodName()); err != nil {
			logger.Errorf("failed to store instance id: %v", err)
		}

		return tracker, nil

	} else if err != nil {
		// unknown error
		return nil, err
	}

	return s.AddressUpdater.TrackAndSyncUpdate(wsApp)
}

func (s *AppSpawner) IsRunning(ws *entity.Workspace, app *entity.ContainerApp) (bool, error) {
	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: ws}
	podName := wsApp.PodName()
	pod, err := s.getPod(podName)
	if err != nil {
		return false, err
	}
	if pod.Status.Phase == "Running" {
		s.AddressUpdater.UpdateFromPod(wsApp, pod)
		return true, nil
	}
	return false, nil
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

	s.AddressUpdater.Reset(wsApp)

	var session = s.mongo.NewSession()
	defer session.Close()
	if err := workspace.RemoveInstances(session, ws.ID, wsApp.PodName()); err != nil {
		logger.Errorf("failed to remove instance id: %v", err)
	}

	// We found the pod, let's start a tracker first, and then delete the pod
	tracker, err := s.AddressUpdater.TrackAndSyncDelete(wsApp)
	if err != nil {
		return nil, err
	}

	var podName = wsApp.PodName()
	var gracePeriodSeconds int64 = 0

	if err := s.clientset.CoreV1().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		defer tracker.Stop()
		if kerrors.IsNotFound(err) {
			return nil, ErrAlreadyStopped
		}
		return nil, err
	}

	return tracker, nil
}

func (s *AppSpawner) getPod(name string) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(name, metav1.GetOptions{})
}
