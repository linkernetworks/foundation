package workspacefsspawner

import (
	"errors"
	"sync"

	"bitbucket.org/linkernetworks/aurora/src/logger"

	_ "bitbucket.org/linkernetworks/aurora/src/aurora"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
	kvolume "bitbucket.org/linkernetworks/aurora/src/kubernetes/volume"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/volumechecker"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"bitbucket.org/linkernetworks/aurora/src/workspace"
	"bitbucket.org/linkernetworks/aurora/src/workspace/fileserver"

	//FIXME, wait PR#444
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"gopkg.in/mgo.v2/bson"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrAlreadyStopped = errors.New("WorkspaceFileServer is already stopped")

type WorkspacePodDeployment interface {
	entity.PodDeployment
	entity.ProxyInfoProvider
}

type WorkspaceFileServerSpawner struct {
	Config config.Config
	Mongo  *mongo.Service

	Updater   *podproxy.ProxyAddressUpdater
	clientset *kubernetes.Clientset
	namespace string
}

func New(c config.Config, m *mongo.Service, clientset *kubernetes.Clientset, rds *redis.Service) *WorkspaceFileServerSpawner {
	return &WorkspaceFileServerSpawner{
		Config:    c,
		Mongo:     m,
		namespace: "default",
		clientset: clientset,
		Updater: &podproxy.ProxyAddressUpdater{
			Clientset: clientset,
			Namespace: "default",
			Redis:     rds,
			PortName:  fileserver.FileServerPortName,
		},
	}
}

func (s *WorkspaceFileServerSpawner) getPod(doc types.DeploymentIDProvider) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

func (s *WorkspaceFileServerSpawner) WakeUp(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	_, err = s.getPod(ws)
	if kerrors.IsNotFound(err) {
		podFactory := fileserver.NewPodFactory(ws)

		pod := podFactory.NewPod(ws.DeploymentID(), map[string]string{
			"service":   "workspace",
			"user":      ws.Owner.Hex(),
			"workspace": ws.ID.Hex(),
		})

		// attach the primary volumes to the pod spec
		if err := workspace.AttachVolumesToPod(ws, &pod); err != nil {
			return nil, err
		}

		tracker, err = s.Updater.TrackAndSyncUpdate(ws)
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
		return nil, err
	}

	return tracker, err
}

func (s *WorkspaceFileServerSpawner) Start(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	podFactory := fileserver.NewPodFactory(ws)

	pod := podFactory.NewPod(ws.DeploymentID(), map[string]string{
		"service": "workspce-fs",
		"user":    ws.Owner.Hex(),
	})

	// attach the primary volumes to the pod spec
	if err := workspace.AttachVolumesToPod(ws, &pod); err != nil {
		return nil, err
	}

	tracker, err = s.Updater.TrackAndSyncUpdate(ws)
	if err != nil {
		return nil, err
	}

	_, err = s.clientset.CoreV1().Pods(s.namespace).Create(&pod)
	if err != nil {
		tracker.Stop()
		return nil, err
	}

	return tracker, nil
}

func (s *WorkspaceFileServerSpawner) Stop(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	// if it's not created
	_, err = s.getPod(ws)
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	s.Updater.Reset(ws)

	// We found the pod, let's start a tracker first, and then delete the pod
	tracker, err = s.Updater.TrackAndSyncDelete(ws)
	if err != nil {
		return nil, err
	}

	err = s.clientset.CoreV1().Pods(s.namespace).Delete(ws.DeploymentID(), &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		tracker.Stop()
		return nil, ErrAlreadyStopped
	} else if err != nil {
		tracker.Stop()
		return nil, err
	}

	return tracker, nil
}

func (s *WorkspaceFileServerSpawner) Restart(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	//Stop the current worksapce-fs pod
	_, err = s.getPod(ws)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	}

	//If pod exist
	if err == nil {
		//Wait the terminatrion event
		//We should wait the delete event by ourself now.
		//sync := tracker.WaitFor(v1.PodSucceeded)
		//sync.Wait()
		m := sync.Mutex{}
		c := sync.NewCond(&m)
		var stop chan struct{}
		_, controller := kubemon.WatchPods(s.clientset, s.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				pod, ok := obj.(*v1.Pod)
				if !ok {
					return
				}

				if pod.ObjectMeta.Name != ws.DeploymentID() {
					return
				}

				c.L.Lock()
				c.Signal()
				c.L.Unlock()
			},
		})

		c.L.Lock()
		go controller.Run(stop)
		tracker, err = s.Stop(ws)
		if err != nil && err != ErrAlreadyStopped {
			c.Signal()
			return nil, err
		}

		logger.Info("Wait for pod=", ws.DeploymentID())
		c.Wait()
		c.L.Unlock()
		logger.Infof("pod=%s has beend deleted", ws.DeploymentID())
	}

	//Start the new fileserver.fs with new config
	logger.Info("Start the pod=%s", ws.DeploymentID())
	tracker, err = s.Start(ws)
	if err != nil {
		return nil, err
	}

	q := bson.M{"_id": ws.GetID()}
	m := bson.M{
		"$set": bson.M{
			"secondaryVolumes": ws.SecondaryVolumes,
		},
	}

	session := s.Mongo.NewSession()
	defer session.Close()
	session.C(entity.WorkspaceCollectionName).Update(q, m)
	return tracker, nil
}

func (s *WorkspaceFileServerSpawner) CheckAvailability(id string, volume *container.Volume, timeout int) error {
	//Deploy a Check POD
	if volume == nil {
		return nil
	}

	pod := volumechecker.NewVolumeCheckPod(id)
	kvolume.AttachVolumeToPod(volume, &pod)
	newPod, err := s.clientset.CoreV1().Pods(s.namespace).Create(&pod)
	if err != nil {
		return err
	}

	defer s.clientset.CoreV1().Pods(s.namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
	//Wait the POD
	o := make(chan *v1.Pod)
	stop := make(chan struct{})
	defer close(stop)
	_, controller := kubemon.WatchPods(s.clientset, s.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}
			o <- pod
		},
	})
	go controller.Run(stop)

	logger.Info("Try to wait the POD", newPod.ObjectMeta.Name)
	if err := volumechecker.Check(o, newPod.ObjectMeta.Name, timeout); err != nil {
		return err
	}

	return nil
}
