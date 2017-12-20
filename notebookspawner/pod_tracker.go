package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/kubemon"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func trackPod(clientset *kubernetes.Clientset, podName, namespace string) (chan *v1.Pod, chan struct{}) {
	o := make(chan *v1.Pod)
	stop := make(chan struct{})

	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			var pod *v1.Pod
			var ok bool
			if pod, ok = newObj.(*v1.Pod); !ok {
				return
			}
			if podName != pod.ObjectMeta.Name {
				return
			}
			o <- pod
		},
	})

	go controller.Run(stop)
	return o, stop
}
