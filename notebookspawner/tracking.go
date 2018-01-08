package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	// import global logger
	"bitbucket.org/linkernetworks/aurora/src/logger"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func (s *NotebookSpawnerService) startTracking(clientset *kubernetes.Clientset, podName string, nb *entity.Notebook) *podtracker.PodTracker {
	topic := nb.Topic()

	podTracker := podtracker.New(clientset, s.namespace, podName)
	podTracker.Track(func(pod *v1.Pod) bool {
		phase := pod.Status.Phase
		logger.Infof("Tracking notebook pod=%s phase=%s", podName, phase)

		s.Redis.PublishAndSetJSON(topic, event.RecordEvent{
			Type: "record.update",
			Update: &event.RecordUpdateEvent{
				Document: "notebooks",
				Id:       nb.ID.Hex(),
				Record:   nb,
				Setter: map[string]interface{}{
					"pod.ip":    pod.Status.PodIP,
					"pod.phase": phase,
				},
			},
		})

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
