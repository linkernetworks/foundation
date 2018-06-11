package appspawner

import (
	"errors"
	"fmt"
	"os"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment/podfactory"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"github.com/linkernetworks/config"
	"github.com/linkernetworks/logger"
	"github.com/linkernetworks/utils/netutils"

	"github.com/linkernetworks/mongo"
	"github.com/linkernetworks/redis"

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

	env := []v1.EnvVar{
		{
			Name:  "AURORA_BASE_URL",
			Value: os.Getenv("SITE_URL"),
		},
	}
	for i, _ := range pod.Spec.Containers {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, env...)
	}

	// attach the primary volumes to the pod spec
	if err := workspace.AttachVolumesToPod(app.Workspace, pod); err != nil {
		return pod, err
	}

	return pod, nil
}

func (s *AppSpawner) Start(ws *entity.Workspace, appRef *entity.ContainerApp, startOption StartOption) (tracker *podtracker.PodTracker, err error) {
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

		//Check the connect
		if startOption.Wait {
			pod := s.getRunningPod(wsApp, startOption.Timeout)
			if pod == nil {
				return nil, fmt.Errorf("Can't start the application %s in %d seconds", wsApp.PodName(), startOption.Timeout)
			}

			port := &wsApp.Container.Ports[0]
			if err := netutils.CheckNetworkConnectivity(pod.Status.PodIP, int(port.ContainerPort), port.Protocol, startOption.Timeout); err != nil {
				return nil, fmt.Errorf("Can't connect to %s: %v", wsApp.PodName(), err)
			}
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

func (s *AppSpawner) getRunningPod(wsApp *entity.WorkspaceApp, timeout int) *v1.Pod {
	//Check is running
	o := make(chan *v1.Pod)
	var stop chan struct{}

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

	stop = make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	var pod *v1.Pod
	pod = nil
	ticker := time.NewTicker(time.Duration(timeout) * time.Second)
Watch:
	for {
		select {
		case <-ticker.C:
			break Watch
		case p := <-o:
			if v1.PodRunning == p.Status.Phase {
				pod = p
				ticker.Stop()
				break Watch
			}
		}
	}

	return pod
}

func (s *AppSpawner) CheckConnectivity(ws *entity.Workspace, appRef *entity.ContainerApp, timeout int) (bool, error) {
	//If not running, return false
	run, err := s.IsRunning(ws, appRef)
	if err != nil {
		return false, err
	}
	if !run {
		return false, nil
	}

	//Get the Pod Object
	app := appRef.Copy()
	wsApp := &entity.WorkspaceApp{ContainerApp: &app, Workspace: ws}
	pod, err := s.getPod(wsApp.PodName())
	if err != nil {
		return false, err
	}

	//Get the protocol from the wsApp
	port := &wsApp.Container.Ports[0]
	if err := netutils.CheckNetworkConnectivity(pod.Status.PodIP, int(port.ContainerPort), port.Protocol, timeout); err != nil {
		return false, fmt.Errorf("Can't connect to %s: %v", wsApp.PodName(), err)
	}

	return true, nil
}
