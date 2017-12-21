package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/kubemon"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodTracker struct {
	clientset *kubernetes.Clientset
	namespace string
	stop      chan struct{}
}

type PodReceiver func(pod *v1.Pod) bool

func NewPodTracker(clientset *kubernetes.Clientset, namespace string) *PodTracker {
	return &PodTracker{clientset, namespace, make(chan struct{})}
}

func matchPod(obj interface{}, podName string) (bool, *v1.Pod) {
	pod, ok := obj.(*v1.Pod)
	return ok && podName == pod.ObjectMeta.Name, pod
}

func (t *PodTracker) Track(podName string, callback PodReceiver) {
	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPod(newObj); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPod(obj); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) Stop() {
	t.stop <- true
	close(t.stop)
}
