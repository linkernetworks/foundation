package appspawner

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment/podfactory"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"bitbucket.org/linkernetworks/aurora/src/workspace"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

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
			"webapp":     &podfactory.GenericPodFactory{},
			"fileserver": &podfactory.GenericPodFactory{},
			"console":    &podfactory.GenericPodFactory{},
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

func (s *AppSpawner) Start(ws *entity.Workspace, appRef *entity.ContainerApp) (tracker *podtracker.PodTracker, err error) {
	app := appRef.Copy()
	wsApp := &entity.WorkspaceApp{ContainerApp: &app, Workspace: ws}
	if appRef.UseEnvironmentImage {
		if image := ws.GetCurrentCPUImage(); len(image) > 0 {
			logger.Infof("[appspawner] starting app %s with image %s", app.Identifier, app.Container.Image)
			app.ApplyImage(image)
		}
	}

	// Start tracking first
	runningPod, err := s.getPod(wsApp.PodName())

	if runningPod != nil && err == nil {
		// already exist, we can update the information from the pod
		s.AddressUpdater.UpdateFromPod(wsApp, runningPod)
		return nil, nil
	}

	if err != nil && kerrors.IsNotFound(err) {

		pod, err := s.NewPod(wsApp)
		if err != nil {
			return nil, err
		}

		// Pod not found. Start a pod of an app in workspace
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
	}

	// unknown error
	return nil, err
}

func (s *AppSpawner) IsRunning(ws *entity.Workspace, appRef *entity.ContainerApp) (bool, error) {
	app := appRef.Copy()
	wsApp := &entity.WorkspaceApp{ContainerApp: &app, Workspace: ws}
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
func (s *AppSpawner) Stop(ws *entity.Workspace, appRef *entity.ContainerApp) (*podtracker.PodTracker, error) {
	app := appRef.Copy()
	wsApp := &entity.WorkspaceApp{ContainerApp: &app, Workspace: ws}

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
	var gracePeriodSeconds int64 = 1
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

func (s *AppSpawner) checkAppIsRunning(wsApp *entity.WorkspaceApp, timeout int) error {
	//Check is running
	o := make(chan *v1.Pod)
	var stop chan struct{}
	defer close(stop)
	_, controller := kubemon.WatchPods(s.clientset, s.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}
			if pod.ObjectMeta.Name != wsApp.PodName() {
				return
			}
			o <- pod
		},
	})
	go controller.Run(stop)

	return s.checkNetworkConnectivity(o, wsApp, timeout)
}

func (s *AppSpawner) checkNetworkConnectivity(ch chan *v1.Pod, wsApp *entity.WorkspaceApp, timeout int) error {
	var find error
	find = nil
	ticker := time.NewTicker(time.Duration(timeout) * time.Second)
Watch:
	for {
		select {
		case pod := <-ch:
			if v1.PodRunning != pod.Status.Phase {
				continue
			}
			//Check the Connectivity
			for {
				port := &wsApp.Container.Ports[0]
				host := net.JoinHostPort(pod.Status.PodIP, strconv.Itoa(int(port.ContainerPort)))
				if conn, err := net.Dial(port.Protocol, host); err == nil {
					conn.Close()
					break Watch
				}
				time.Sleep(time.Duration(1) * time.Second)
			}
		case <-ticker.C:
			find = fmt.Errorf("AA")
			break Watch
		}
	}

	return find
}
