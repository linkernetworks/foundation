package notebookspawner

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The container port of jupyter notebook
const DefaultNotebookContainerPort = 8888

type NotebookPodParameters struct {
	// Notebook parameters
	WorkDir      string
	WorkspaceDir string
	Image        string
	BaseURL      string
	Bind         string
	Port         int32
	Volumes      []container.Volume
}

// NotebookPodFactory handle the process of creating the jupyter notebook pod
type NotebookPodFactory struct {
	notebook *entity.Notebook
	params   NotebookPodParameters
}

func NewNotebookPodFactory(notebook *entity.Notebook, params NotebookPodParameters) *NotebookPodFactory {
	return &NotebookPodFactory{notebook, params}
}

func NewVolume(params PodParameters) []v1.Volume {
	kubeVolume := []v1.Volume{}
	for _, v := range params.Volumes {
		kubeVolume = append(kubeVolume, v1.Volume{
			Name: v.VolumeMount.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.ClaimName,
				},
			},
		})
	}
	kubeVolume = append(kubeVolume, v1.Volime{
		Name: "config-volume",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "jupyter-notebook-config",
				},
			},
		},
	})

	return kubeVolume
}

func NewVolumeMount(params PodParameters) []v1.VolumeMount {
	kubeVolumeMount := []v1.VolumeMount{}
	for _, v := range params.Volumes {
		kubeVolumeMount = append(kubeVolumeMount, v1.VolumeMount{
			Name:      v.VolumeMount.Name,
			SubPath:   v.VolumeMount.SubPath,
			MountPath: v.VolumeMount.MountPath,
		})
	}

	kubeVoumeMount = append(kubeVolumeMount, v1.VolumeMount{
		Name:      "config-volume",
		MountPath: "/home/jovyan/.jupyter/custom",
	})
	return kubeVolumeMount
}

// NewPod returns the Pod object of the jupyternotebook
func (nb *NotebookPodFactory) NewPod(podName string, labels map[string]string) v1.Pod {
	params := nb.params
	kubeVolume := NewVolume(params)
	kubeVolumeMount := NewVolumeMount(params)

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: labels,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "notebook",
					Image:           nb.params.Image,
					ImagePullPolicy: v1.PullIfNotPresent,
					Args: []string{
						"start-notebook.sh",
						"--notebook-dir=" + nb.params.WorkDir,
						"--ip=" + nb.params.Bind,
						"--port=" + strconv.Itoa(int(nb.params.Port)),
						"--NotebookApp.base_url=" + nb.params.BaseURL,
						"--NotebookApp.token=''",
						"--NotebookApp.allow_origin='*'",
						"--NotebookApp.disable_check_xsrf=True",
						"--Session.debug=True",
					},
					//FIXME we should also mount the PrimaryVolume.
					VolumeMounts: kubeVolumeMount,
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
			Volumes: kubeVolume,
		},
	}
}
