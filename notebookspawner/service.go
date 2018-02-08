package notebookspawner

import (
	"errors"

	"bitbucket.org/linkernetworks/aurora/src/aurora/provision/path"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/types"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	kubernetesclient "k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"gopkg.in/mgo.v2/bson"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SpawnableDocument interface {
	types.DeploymentIDProvider
	GetID() bson.ObjectId
	Topic() string
	NewUpdateEvent(info bson.M) *event.RecordEvent
}

type DocumentProxyInfoUpdater struct {
	clientset *kubernetesclient.Clientset
	namespace string

	Redis          *redis.Service
	Session        *mongo.Session
	CollectionName string

	// The PortName
	PortName string
}

func (u *DocumentProxyInfoUpdater) getPod(doc SpawnableDocument) (*v1.Pod, error) {
	return u.clientset.CoreV1().Pods(u.namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

// TrackAndSync tracks the pod of the owner document and returns a pod tracker
func (u *DocumentProxyInfoUpdater) TrackAndSync(doc SpawnableDocument) (*podtracker.PodTracker, error) {
	podName := doc.DeploymentID()

	podTracker := podtracker.New(u.clientset, u.namespace, podName)

	podTracker.Track(func(pod *v1.Pod) bool {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		switch phase {
		case "Pending":
			u.SyncWithPod(doc, pod)
			// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
			for _, c := range pod.Status.ContainerStatuses {
				if c.State.Waiting != nil {
					waitingReason := c.State.Waiting.Reason
					if waitingReason == "ErrImagePull" || waitingReason == "ImagePullBackOff" {
						logger.Errorf("Container is waiting. Reason %s\n", waitingReason)

						// stop tracking
						return true
					}
				}
			}

		// Stop the tracker if the status is completion status.
		// Terminating won't be catched
		case "Running", "Failed", "Succeeded", "Unknown", "Terminating":
			u.SyncWithPod(doc, pod)
			return true
		}

		return false
	})
	return podTracker, nil
}

func (u *DocumentProxyInfoUpdater) Sync(doc SpawnableDocument) error {
	pod, err := u.getPod(doc)

	if err != nil && kerrors.IsNotFound(err) {

		return u.Reset(doc, nil)

	} else if err != nil {

		u.Reset(doc, err)
		return err
	}

	return u.SyncWithPod(doc, pod)
}

func (u *DocumentProxyInfoUpdater) Reset(doc SpawnableDocument, kerr error) (err error) {
	var q = bson.M{"_id": doc.GetID()}
	var m = bson.M{
		"$set": bson.M{
			"backend.connected": false,
			"backend.error":     kerr,
		},
		"$unset": bson.M{
			"backend.host": nil,
			"backend.port": nil,
			"pod":          nil,
		},
	}
	err = u.Session.C(u.CollectionName).Update(q, m)
	u.emit(doc, doc.NewUpdateEvent(bson.M{
		"backend.connected": false,
		"backend.host":      nil,
		"backend.port":      nil,
		"pod":               nil,
	}))
	return err
}

// SyncWith updates the given document's "backend" and "pod" field by the given
// pod object.
func (p *DocumentProxyInfoUpdater) SyncWithPod(doc SpawnableDocument, pod *v1.Pod) (err error) {
	backend, err := podproxy.NewProxyBackendFromPodStatus(pod, p.PortName)
	if err != nil {
		return err
	}

	q := bson.M{"_id": doc.GetID()}
	m := bson.M{
		"$set": bson.M{
			"backend": backend,
			"pod":     podproxy.NewPodInfo(pod),
		},
	}

	if err = p.Session.C(p.CollectionName).Update(q, m); err != nil {
		return err
	}

	p.emit(doc, doc.NewUpdateEvent(bson.M{
		"backend.connected": pod.Status.PodIP != "",
		"pod.phase":         pod.Status.Phase,
		"pod.message":       pod.Status.Message,
		"pod.reason":        pod.Status.Reason,
		"pod.startTime":     pod.Status.StartTime,
	}))
	return nil
}

func (p *DocumentProxyInfoUpdater) emit(doc SpawnableDocument, e *event.RecordEvent) {
	go p.Redis.PublishAndSetJSON(doc.Topic(), e)
}

var ErrAlreadyStopped = errors.New("Notebook is already stopped")

type NotebookSpawnerService struct {
	Config  config.Config
	Session *mongo.Session

	updater *DocumentProxyInfoUpdater

	clientset *kubernetesclient.Clientset
	namespace string
}

func New(c config.Config, m *mongo.Service, clientset *kubernetesclient.Clientset, rds *redis.Service) *NotebookSpawnerService {
	session := m.NewSession()
	return &NotebookSpawnerService{
		Config:    c,
		Session:   session,
		namespace: "default",
		clientset: clientset,
		updater: &DocumentProxyInfoUpdater{
			clientset:      clientset,
			namespace:      "default",
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

	// volumeMounts subPath should not have a root dir. the correct one is like batches/batch-xxx
	pvSubpath := path.GetWorkspacePVSubpath(s.Config, &workspace)

	podFactory := NewNotebookPodFactory(nb, NotebookPodParameters{
		Image:        nb.Image,
		WorkspaceDir: pvSubpath,
		WorkingDir:   s.Config.Jupyter.WorkingDir,
		Bind:         s.Config.Jupyter.Address,
		Port:         NotebookContainerPort,
		BaseURL:      nb.Url,
	})

	pod := podFactory.NewPod(podName, map[string]string{
		"service":   "notebook",
		"workspace": nb.WorkspaceID.Hex(),
		"user":      nb.CreatedBy.Hex(),
	})

	// Start tracking first
	_, err = s.getPod(nb)
	if kerrors.IsNotFound(err) {
		// Pod not found. Start a pod for notebook in workspace(batch)
		tracker, err = s.updater.TrackAndSync(nb)
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

	tracker, err = s.updater.TrackAndSync(nb)
	return tracker, err
}

func (s *NotebookSpawnerService) getPod(doc types.DeploymentIDProvider) (*v1.Pod, error) {
	return s.clientset.CoreV1().Pods(s.namespace).Get(doc.DeploymentID(), metav1.GetOptions{})
}

// Stop returns nil if it's already stopped
func (s *NotebookSpawnerService) Stop(nb *entity.Notebook) (*podtracker.PodTracker, error) {
	// if it's not created
	_, err := s.getPod(nb)
	if kerrors.IsNotFound(err) {
		return nil, ErrAlreadyStopped
	} else if err != nil {
		return nil, err
	}

	podName := nb.DeploymentID()

	// force sending a terminating state to document
	q := bson.M{"_id": nb.GetID()}
	m := bson.M{
		"$set": bson.M{
			"backend.connected": false,
			"pod.phase":         "Terminating",
		},
	}
	s.Session.C(entity.NotebookCollectionName).Update(q, m)

	// We found the pod, let's start a tracker first, and then delete the pod
	podTracker, err := s.updater.TrackAndSync(nb)
	if err != nil {
		return nil, err
	}

	err = s.clientset.CoreV1().Pods(s.namespace).Delete(podName, &metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		podTracker.Stop()
		return nil, ErrAlreadyStopped
	} else if err != nil {
		podTracker.Stop()
		return nil, err
	}
	return podTracker, nil
}
