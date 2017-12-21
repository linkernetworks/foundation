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
	C         chan *v1.Pod
	stop      chan struct{}
}

func NewPodTracker(clientset *kubernetes.Clientset, namespace string) *PodTracker {
	return &PodTracker{clientset, namespace, make(chan *v1.Pod), make(chan struct{})}
}

func matchPod(obj interface{}, podName string) (bool, *v1.Pod) {
	pod, ok := newObj.(*v1.Pod)
	return ok && podName == pod.ObjectMeta.Name, pod
}

func (t *PodTracker) Track(podName string) chan *v1.Pod {
	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			if pod, ok := matchPod(newObj) ; ok {
				t.C <- pod
			}
		},
		DeleteFunc: func(obj interface{}) {
			if pod, ok := matchPod(obj) ; ok {
				t.C <- pod
			}
		},
	})

	go controller.Run(t.stop)
	return t.C
}

func (t *PodTracker) Stop() {
	t.stop<-
}
