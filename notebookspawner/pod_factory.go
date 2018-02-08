package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NotebookContainerPort = 8888

// NotebookPodParameters stores the parameters that will be used for creating
// the notebook pod.
type NotebookPodParameters struct {
	// Notebook parameters
	WorkingDir   string
	WorkspaceDir string
	Image        string
	BaseURL      string
	Bind         string
	Port         int32
}

// NotebookPodFactory handle the process of creating the jupyter notebook pod
type NotebookPodFactory struct {
	notebook *entity.Notebook
	params   NotebookPodParameters
}

func NewNotebookPodFactory(notebook *entity.Notebook, params NotebookPodParameters) *NotebookPodFactory {
	return &NotebookPodFactory{notebook, params}
}

// DeploymentID returns the name of the deployment.
// For Kubernetes, it will be the pod name.
func (nb *NotebookPodFactory) DeploymentID() string {
	return nb.notebook.ID.Hex()
}

// NewPod returns the Pod object of the jupyternotebook
func (nb *NotebookPodFactory) NewPod(podName string, labels map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: labels,
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers: []v1.Container{
				{
					Image:           nb.params.Image,
					Name:            podName,
					ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
					Args: []string{
						"start-notebook.sh",
						"--notebook-dir=" + nb.params.WorkingDir,
						"--ip=" + nb.params.Bind,
						"--port=" + strconv.Itoa(int(nb.params.Port)),
						"--NotebookApp.base_url=" + nb.params.BaseURL,
						"--NotebookApp.token=''",
						"--NotebookApp.allow_origin='*'",
						"--NotebookApp.disable_check_xsrf=True",
						"--Session.debug=True",
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "data-volume",
							SubPath:   nb.params.WorkspaceDir,
							MountPath: nb.params.WorkingDir,
						},
						{Name: "config-volume", MountPath: "/home/jovyan/.jupyter/custom"},
					},
					Ports: []v1.ContainerPort{
						{
							Name:          "notebook",
							ContainerPort: nb.params.Port,
							Protocol:      v1.ProtocolTCP,
						},
					},
					Env: []v1.EnvVar{
						{
							Name:  "CPU_GUARANTEE",
							Value: "50m",
						},
						{
							Name:  "MEM_GUARANTEE",
							Value: "64Mi",
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{"cpu": resource.MustParse("1000m")},
						Requests: v1.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
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
				{
					Name: "config-volume",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "jupyter-notebook-config",
							},
						},
					},
				},
			},
		},
	}
}
