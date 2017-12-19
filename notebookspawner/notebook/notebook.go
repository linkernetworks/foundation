package notebook

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NotebookContainerPort = 8888
	NotebookPodNamePrefix = "pod-"
)

type KubeNotebook struct {
	Name      string
	Workspace string
	ProxyURL  string
	Image     string
}

func (nb *KubeNotebook) GetPodName() string {
	return NotebookPodNamePrefix + nb.Name
}

func (nb *KubeNotebook) NewPod() v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: nb.GetPodName(),
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers:    nb.Containers(),
			Volumes: []v1.Volume{
				{
					Name: "data-volume",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "data-storage",
						},
					},
				},
			},
		},
	}
}

func (nb *KubeNotebook) Containers() []v1.Container {
	return []v1.Container{
		{
			Image:           nb.Image,
			Name:            nb.GetPodName(),
			ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
			Args: []string{
				"start-notebook.sh",
				"--notebook-dir=/batch",
				"--ip=\"0.0.0.0\"",
				"--port=" + strconv.Itoa(NotebookContainerPort),
				"--NotebookApp.base_url=" + nb.ProxyURL + "/" + nb.Name,
				"--NotebookApp.token=''",
				"--NotebookApp.allow_origin='*'",
				"--NotebookApp.disable_check_xsrf=True",
				"--Session.debug=True",
			},
			VolumeMounts: []v1.VolumeMount{
				{Name: "data-volume", SubPath: nb.Workspace, MountPath: "/batch"},
			},
			Ports: []v1.ContainerPort{
				{ContainerPort: NotebookContainerPort, Name: "notebook-port", Protocol: v1.ProtocolTCP},
			},
			Env: []v1.EnvVar{
				{
					Name:  "CPU_GUARANTEE",
					Value: "200m",
				},
				{
					Name:  "MEM_GUARANTEE",
					Value: "256Mi",
				},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"cpu":    resource.MustParse("200m"),
					"memory": resource.MustParse("256Mi"),
				},
			},
		},
	}
}
