package appspawner

import (
	"errors"
	"net"
	"strconv"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podutil"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	"k8s.io/client-go/kubernetes"

	kerrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrPortNotFound = errors.New("Container port not found.")

func NewPodInfo(pod *v1.Pod) *entity.PodInfo {
	return &entity.PodInfo{
		Phase:     podproxy.HandlePodPhase(pod),
		Message:   pod.Status.Message,
		Reason:    pod.Status.Reason,
		StartTime: pod.Status.StartTime,
	}
}

type ProxyAddressUpdater struct {
	Clientset *kubernetes.Clientset
	Namespace string

	Cache *podproxy.ProxyCache
}

func (u *ProxyAddressUpdater) getPod(app *entity.WorkspaceApp) (*v1.Pod, error) {
	return u.Clientset.CoreV1().Pods(u.Namespace).Get(app.DeploymentID(), metav1.GetOptions{})
}

// NewSyncHandler returns a function that handles the pod changes received from the pod tracker.
//
// The following comments are copied from the kubernetes repository:
//
//     PodPending means the pod has been accepted by the system, but one or more of the containers
//     has not been started. This includes time before being bound to a node, as well as time spent
//     pulling images onto the host.
//
//    		PodPending PodPhase = "Pending"
//
//     PodRunning means the pod has been bound to a node and all of the containers have been started.
//     At least one container is still running or is in the process of being restarted.
//
//    		PodRunning PodPhase = "Running"
//
//     PodSucceeded means that all containers in the pod have voluntarily terminated
//     with a container exit code of 0, and the system is not going to restart any of these containers.
//
//    		PodSucceeded PodPhase = "Succeeded"
//
//     PodFailed means that all containers in the pod have terminated, and at least one container has
//     terminated in a failure (exited with a non-zero exit code or was stopped by the system).
//
//    		PodFailed PodPhase = "Failed"
//
//     PodUnknown means that for some reason the state of the pod could not be obtained, typically due
//     to an error in communicating with the host of the pod.
//
//    		PodUnknown PodPhase = "Unknown"
//
// See package "k8s.io/kubernetes/pkg/apis/core/types.go" for more details.
func (u *ProxyAddressUpdater) NewSyncHandler(app *entity.WorkspaceApp) func(pod *v1.Pod) (stop bool) {
	podName := app.DeploymentID()

	return func(pod *v1.Pod) (stop bool) {
		phase := pod.Status.Phase
		logger.Infof("podproxy: found change: pod=%s phase=%s", podName, phase)

		switch phase {
		case v1.PodPending:
			if err := u.UpdateFromPod(app, pod); err != nil {
				logger.Errorf("podproxy: failed to sync address: pod=%s error=%v", podName, err)
			}

			// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
			cslist := podutil.FindWaitingContainerStatuses(pod)
			for _, cs := range cslist {
				reason := cs.State.Waiting.Reason
				switch reason {
				case "PodInitializing", "ContainerCreating":
					// Skip the standard states
					logger.Infof("podproxy: container state %s", reason)

				case "ErrImageInspect",
					"ErrImagePullBackOff",
					"ErrImagePull",
					"ErrImageNeverPull",
					"RegistryUnavailable",
					"ErrInvalidImageName":
					logger.Errorf("podproxy: container is waiting: reason=%s", cs.ContainerID, reason)

					// stop tracking
					stop = true
					return stop

				default:
					logger.Errorf("podproxy: unexpected reason=%s", reason)

				}
			}

		// Stop the tracker if the status is completion status.
		// Terminating won't be catched
		case v1.PodRunning, v1.PodFailed, v1.PodSucceeded, v1.PodUnknown:
			if err := u.UpdateFromPod(app, pod); err != nil {
				logger.Errorf("podproxy: failed to sync document: pod=%s error=%v", podName, err)
			}

			stop = true
			return stop
		default:
			logger.Errorf("podproxy: phase %s not handled.", phase)
		}

		stop = false
		return stop
	}
}

func (u *ProxyAddressUpdater) TrackAndSyncAdd(app *entity.WorkspaceApp) (*podtracker.PodTracker, error) {
	podName := app.DeploymentID()

	tracker := podtracker.New(u.Clientset, u.Namespace, podName)

	tracker.TrackAdd(u.NewSyncHandler(app))
	return tracker, nil
}

func (u *ProxyAddressUpdater) TrackAndSyncUpdate(app *entity.WorkspaceApp) (*podtracker.PodTracker, error) {
	podName := app.DeploymentID()
	tracker := podtracker.New(u.Clientset, u.Namespace, podName)
	tracker.TrackUpdate(u.NewSyncHandler(app))
	return tracker, nil
}

func (u *ProxyAddressUpdater) TrackAndSyncDelete(app *entity.WorkspaceApp) (*podtracker.PodTracker, error) {
	podName := app.DeploymentID()

	tracker := podtracker.New(u.Clientset, u.Namespace, podName)

	tracker.TrackDelete(u.NewSyncHandler(app))
	return tracker, nil
}

func (u *ProxyAddressUpdater) Sync(app *entity.WorkspaceApp) error {
	pod, err := u.getPod(app)

	if err != nil && kerrors.IsNotFound(err) {

		return u.Reset(app)

	} else if err != nil {

		u.Reset(app)
		return err
	}

	return u.UpdateFromPod(app, pod)
}

func (u *ProxyAddressUpdater) Reset(app *entity.WorkspaceApp) error {
	return u.Cache.RemoveAddress(app.DeploymentID())
}

// SyncWith updates the given document's "backend" and "pod" field by the given
// pod object.
func (u *ProxyAddressUpdater) UpdateFromPod(app *entity.WorkspaceApp, pod *v1.Pod) (err error) {
	if len(app.Container.ExposePortName) > 0 {
		logger.Debugf("podproxy: syncing document proxy info: %s", app.DeploymentID())
		port, ok := podutil.FindContainerPort(pod, app.Container.ExposePortName)
		if !ok {
			return ErrPortNotFound
		}

		return u.Cache.SetAddress(app.DeploymentID(), net.JoinHostPort(pod.Status.PodIP, strconv.Itoa(port)))
	}

	return nil
}
