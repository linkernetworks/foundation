package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	// import global logger
	"bitbucket.org/linkernetworks/aurora/src/logger"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func (s *NotebookSpawnerService) startTracking(clientset *kubernetes.Clientset, podName string, nb *entity.Notebook) *podtracker.PodTracker {
	podTracker := podtracker.New(clientset, s.namespace, podName)
	podTracker.Track(func(pod *v1.Pod) bool {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		switch phase {
		case "Pending":
			s.Sync(nb.ID, pod)
			// Check all containers status in a pod. can't be ErrImagePull or ImagePullBackOff
			for _, c := range pod.Status.ContainerStatuses {
				waitingReason := c.State.Waiting.Reason
				if waitingReason == "ErrImagePull" || waitingReason == "ImagePullBackOff" {
					logger.Errorf("Container is waiting. Reason %s\n", waitingReason)

					// stop tracking
					return true
				}
			}

		case "Running", "Failed", "Succeeded", "Unknown", "Terminating":
			s.Sync(nb.ID, pod)

			// stop tracking
			return true
		}

		return false
	})
	return podTracker
}

func (s *NotebookSpawnerService) stopTracking(clientset *kubernetes.Clientset, podName string) {
	podTracker := podtracker.New(clientset, s.namespace, podName)
	podTracker.Stop()
	return nil
}
