package notebookspawner

import (
	"strconv"

	"bitbucket.org/linkernetworks/aurora/src/entity"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

// The container port of jupyter notebook
const DefaultNotebookContainerPort = 8888

type NotebookPodParameters struct {
	WorkDir string
	Bind    string
	Port    int32
}

// NotebookPodFactory handle the process of creating the jupyter notebook pod
type NotebookPodFactory struct {
	params NotebookPodParameters
}

func NewNotebookPodFactory(params NotebookPodParameters) *NotebookPodFactory {
	return &NotebookPodFactory{params}
}

// NewPod returns the Pod object of the jupyternotebook
func (nb *NotebookPodFactory) NewPod(notebook *entity.Notebook) v1.Pod {
	podName := notebook.DeploymentID()
	volumes := []v1.Volume{
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
	}

	mounts := []v1.VolumeMount{
		{
			Name:      "config-volume",
			MountPath: "/home/jovyan/.jupyter/custom",
		},
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			Labels: map[string]string{
				"service":   "notebook",
				"workspace": notebook.WorkspaceID.Hex(),
				"user":      notebook.CreatedBy.Hex(),
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "notebook",
					Image:           notebook.Image,
					ImagePullPolicy: v1.PullIfNotPresent,
					Args: []string{
						"start-notebook.sh",
						"--notebook-dir=" + nb.params.WorkDir,
						"--ip=" + nb.params.Bind,
						"--port=" + strconv.Itoa(int(nb.params.Port)),
						"--NotebookApp.base_url=" + notebook.Url,
						"--NotebookApp.token=''",
						"--NotebookApp.allow_origin='*'",
						"--NotebookApp.disable_check_xsrf=True",
						"--Session.debug=True",
					},
					//FIXME we should also mount the PrimaryVolume.
					VolumeMounts: mounts,
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
			Volumes: volumes,
		},
	}
}
