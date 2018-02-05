package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const FileserverContainerPort = 8888

type FileserverPodFactory struct {
	Fileserver *entity.Fileserver
}

type FileserverPodParameters struct {
	// Fileserver parameters
	Port   int32
	Image  string
	Labels map[string]string
	// TODO we need to user another structure for volumes
	// That structur should including all kinds of storage (NFS,Local,GlusterFS)
	// The current src/types/containers/types' Volumes only for volumeMounts
}

func (nb *FileserverPodFactory) DeploymentID() string {
	return nb.Fileserver.ID.Hex()
}

func (nb *FileserverPodFactory) NewPod(podName string, params FileserverPodParameters) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: params.Labels,
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Always",
			Containers: []v1.Container{
				{
					Image:           params.Image,
					Name:            podName,
					ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
					Args: []string{
						"-p " + strconv.Itoa(int(params.Port)),
					},
					//TODO Add mounts
					VolumeMounts: []v1.VolumeMount{},
					Ports: []v1.ContainerPort{
						{
							Name:          "fileserver",
							ContainerPort: params.Port,
							Protocol:      v1.ProtocolTCP,
						},
					},
				},
			},
			//TODO Add mounts
			Volumes: []v1.Volume{},
		},
	}
}
