package workspacefsspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"errors"
	"sync"

	_ "bitbucket.org/linkernetworks/aurora/src/aurora"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"bitbucket.org/linkernetworks/aurora/src/workspace/fileserver"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	//FIXME, wait PR#444
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	"gopkg.in/mgo.v2/bson"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrDoesNotExist = errors.New("WorkspaceFileServer isn't exist ")

type WorkspacePodDeployment interface {
	entity.PodDeployment
	entity.ProxyInfoProvider
}

type WorkspaceFileServerSpawner struct {
	Config  config.Config
	Session *mongo.Session

	updater   *podproxy.DocumentProxyInfoUpdater
	clientset *kubernetes.Clientset
	namespace string
}

func New(c config.Config, m *mongo.Service, clientset *kubernetes.Clientset, rds *redis.Service) *WorkspaceFileServerSpawner {
	session := m.NewSession()
	return &WorkspaceFileServerSpawner{
		Config:    c,
		Session:   session,
		namespace: "default",
		clientset: clientset,
		updater: &podproxy.DocumentProxyInfoUpdater{
			Clientset:      clientset,
			Namespace:      "default",
			Redis:          rds,
			Session:        session,
			CollectionName: entity.WorkspaceCollectionName,
			PortName:       fileserver.FileServerPortName,
		},
	}
}

func (s *WorkspaceFileServerSpawner) getPod(doc types.DeploymentIDProvider) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

func (s *WorkspaceFileServerSpawner) WakeUp(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	_, err = s.getPod(ws)
	if kerrors.IsNotFound(err) {
		//Create pod
		volumes := []container.Volume{
			{
				ClaimName: ws.PrimaryVolume.Name,
				VolumeMount: container.VolumeMount{
					Name:      ws.PrimaryVolume.Name,
					MountPath: fileserver.MainVolumeMountPoint,
				},
			},
		}

		volumes = append(volumes, ws.SecondaryVolumes...)

		podFactory := fileserver.NewPodFactory(ws, fileserver.PodParameters{
			//FIXME for testing, use develop
			//		Image:   WorkspaceImage + ":" + aurora.ImageTag,
			Image:   fileserver.Image + ":develop",
			Port:    fileserver.ContainerPort,
			Volumes: volumes,
		})

		pod := podFactory.NewPod(ws.DeploymentID(), map[string]string{
			"service": "workspce-fs",
			"user":    ws.Owner.Hex(),
		})

		tracker, err = s.updater.TrackAndSync(ws)
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

func (s *WorkspaceFileServerSpawner) Delete(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	_, err = s.getPod(ws)
	if kerrors.IsNotFound(err) {
		return nil, ErrDoesNotExist
	} else if err != nil {
		return nil, err
	}

	q := bson.M{"_id": ws.GetID()}
	m := bson.M{
		"$set": bson.M{
			"backend.connected": false,
			"pod.phase":         "Terminating",
		},
	}

	s.Session.C(entity.WorkspaceCollectionName).Update(q, m)
	tracker, err = s.updater.TrackAndSync(ws)
	if err != nil {
		return nil, err
	}

	err = s.clientset.CoreV1().Pods(s.namespace).Delete(ws.DeploymentID(), &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		tracker.Stop()
		return nil, ErrDoesNotExist
	} else if err != nil {
		tracker.Stop()
		return nil, err
	}

	return tracker, nil
}

func (s *WorkspaceFileServerSpawner) Restart(ws *entity.Workspace) (tracker *podtracker.PodTracker, err error) {
	//Stop the current worksapce-fs pod
	_, err = s.getPod(ws)
	if err != nil && err != ErrDoesNotExist {
		return nil, err
	}

	if err != ErrDoesNotExist {

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
		tracker, err = s.Delete(ws)
		if err != nil && err != ErrDoesNotExist {
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
	tracker, err = s.WakeUp(ws)
	if err != nil {
		return nil, err
	}

	q := bson.M{"_id": ws.GetID()}
	m := bson.M{
		"$set": bson.M{
			"subVolumes": ws.SecondaryVolumes,
		},
	}
	s.Session.C(entity.WorkspaceCollectionName).Update(q, m)
	return tracker, nil
}

func (s *WorkspaceFileServerSpawner) GetKubeVolume(ws *entity.Workspace) (volumes []container.Volume, err error) {
	volumes = append(volumes, container.Volume{
		ClaimName: ws.PrimaryVolume.Name,
		VolumeMount: container.VolumeMount{
			Name:      ws.PrimaryVolume.Name,
			MountPath: fileserver.MainVolumeMountPoint,
		},
	})

	volumes = append(volumes, ws.SecondaryVolumes...)
	return volumes, nil
}
