package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/event"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/podtracker"

	// import global logger
	"bitbucket.org/linkernetworks/aurora/src/logger"

	v1 "k8s.io/api/core/v1"
)

func (s *NotebookSpawnerService) startTracking(podName string, nb *entity.Notebook) *podtracker.PodTracker {
	topic := nb.Topic()

	podTracker := podtracker.New(s.clientset, s.namespace, podName)
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
					"backend.ip":        pod.Status.PodIP,
					"backend.port":      NotebookContainerPort,
					"backend.connected": pod.Status.PodIP != "",

					"pod.phase":     pod.Status.Phase,
					"pod.message":   pod.Status.Message,
					"pod.reason":    pod.Status.Reason,
					"pod.startTime": pod.Status.StartTime,
				},
			},
		})

		switch phase {
		case "Pending":
			s.SyncFromPod(nb.ID, pod)
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
			s.SyncFromPod(nb.ID, pod)

			// stop tracking
			return true
		}

		return false
	})
	return podTracker
}
