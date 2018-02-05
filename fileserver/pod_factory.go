package fileserver

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const FileServerContainerPort = 8888

type FileServerPodFactory struct {
	FileServer *entity.FileServer
}

type FileServerPodParameters struct {
	// FileServer parameters
	Port   int32
	Image  string
	Labels map[string]string
	// TODO we need to user another structure for volumes
	// That structur should including all kinds of storage (NFS,Local,GlusterFS)
	// The current src/types/containers/types' Volumes only for volumeMounts
}

func (nb *FileServerPodFactory) DeploymentID() string {
	return nb.FileServer.ID.Hex()
}

func (nb *FileServerPodFactory) NewPod(podName string, params FileServerPodParameters) v1.Pod {
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
