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
	podName   string
	stop      chan struct{}
}

type PodReceiver func(pod *v1.Pod) bool

func NewPodTracker(clientset *kubernetes.Clientset, namespace string, podName string) *PodTracker {
	return &PodTracker{clientset, namespace, podName, make(chan struct{})}
}

func matchPod(obj interface{}, podName string) (*v1.Pod, bool) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil, false
	}
	return pod, podName == pod.ObjectMeta.Name
}

func (t *PodTracker) Track(callback PodReceiver) {
	_, controller := kubemon.WatchPods(t.clientset, t.namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPod(newObj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPod(obj, t.podName); ok {
				if callback(pod) {
					t.Stop()
				}
			}
		},
	})
	go controller.Run(t.stop)
}

func (t *PodTracker) Stop() {
	var e struct{}
	t.stop <- e
	close(t.stop)
}
