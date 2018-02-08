package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	// import global logger
	"bitbucket.org/linkernetworks/aurora/src/logger"

	v1 "k8s.io/api/core/v1"
)

func (s *NotebookSpawnerService) startTracking(podName string, collectionName string, doc SpawnableDocument) (*podtracker.PodTracker, error) {
	clientset, err := s.getClientset()
	if err != nil {
		return nil, err
	}

	podTracker := podtracker.New(clientset, s.namespace, podName)
	podTracker.Track(func(pod *v1.Pod) bool {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		switch phase {
		case "Pending":
			s.SyncDocument(collectionName, doc, pod)
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

		case "Running", "Failed", "Succeeded", "Unknown", "Terminating":
			s.SyncDocument(collectionName, doc, pod)
			return true
		}

		return false
	})
	return podTracker, nil
}
