package workspacefsspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/kubemon"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"errors"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const PrefixPodName = "fs-check-"

var ErrMountUnAvailable = errors.New("Volume Unavailable")

func NewVolume(volume []container.Volume) []v1.Volume {
	kubeVolume := []v1.Volume{}
	for _, v := range volume {
		kubeVolume = append(kubeVolume, v1.Volume{
			Name: v.VolumeMount.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.ClaimName,
				},
			},
		})
	}
	return kubeVolume
}

func NewVolumeMount(volume []container.Volume) []v1.VolumeMount {
	kubeVolumeMount := []v1.VolumeMount{}
	for _, v := range volume {
		kubeVolumeMount = append(kubeVolumeMount, v1.VolumeMount{
			Name:      v.VolumeMount.Name,
			MountPath: v.VolumeMount.MountPath,
			SubPath:   v.VolumeMount.SubPath,
		})
	}
	return kubeVolumeMount
}

func NewAvailablePod(id string, volume []container.Volume) v1.Pod {
	kubeVolume := NewVolume(volume)
	kubeVolumeMount := NewVolumeMount(volume)
	name := PrefixPodName + id
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Always",
			Containers: []v1.Container{{
				Image:           "busybox:latest",
				Name:            name,
				ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
				VolumeMounts:    kubeVolumeMount,
				Command:         []string{"sleep", "3600"},
			},
			},
			Volumes: kubeVolume,
		},
	}
}

func WaitAvailiablePod(clientset *kubernetes.Clientset, namespace string, podName string, timeout int) error {
	//We return nil iff the POD's status is running within timeout seconds.
	find := false
	stop := make(chan struct{})
	_, controller := kubemon.WatchPods(clientset, namespace, fields.Everything(), cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}

			if pod.ObjectMeta.Name != podName {
				return
			}

			if v1.PodRunning == pod.Status.Phase {
				find = true
				var e struct{}
				stop <- e
			}
		},
	})
	go controller.Run(stop)

	//Wait the check unti the timeout
	for i := 0; i < timeout; i++ {
		if find {
			break
		}
		time.Sleep(1 * time.Second)
	}

	var e struct{}
	stop <- e

	if !find {
		return ErrMountUnAvailable
	}
	return nil
}
