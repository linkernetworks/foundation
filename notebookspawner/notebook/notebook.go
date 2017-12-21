package notebook

import (
	"strconv"

	"bitbucket.org/linkernetworks/aurora/src/entity"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NotebookContainerPort = 8888
	NotebookPodNamePrefix = "pod-"
)

type KubeNotebook struct {
	Notebook  *entity.Notebook
	Name      string
	Workspace string
	ProxyURL  string
	Image     string
}

func (nb *KubeNotebook) DeploymentID() string {
	return nb.Notebook.ID.Hex()
}

func (nb *KubeNotebook) NewPod(podName string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers: []v1.Container{
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
			},
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

// Backward compatible method
func (nb *KubeNotebook) Pod() v1.Pod {
	return nb.NewPod(nb.GetPodName())
}

func (nb *KubeNotebook) GetPodName() string {
	return NotebookPodNamePrefix + nb.Name
}
