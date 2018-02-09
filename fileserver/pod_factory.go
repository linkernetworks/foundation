package workspace

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const WorkspacePodNamePrefix = "workspace-fs-"
const WorkspaceContainerPort = 33333
const WorkspaceImage = "asia.gcr.io/linker-aurora/file-server"

type WorkspacePodParameters struct {
	// Workspace parameters
	Port    int32
	Image   string
	Labels  map[string]string
	Volumes []container.Volume
}

type WorkspacePodFactory struct {
	workspace *entity.Workspace
	params    WorkspacePodParameters
}

func NewWorkspacePodFactory(workspace *entity.Workspace, params WorkspacePodParameters) *WorkspacePodFactory {
	return &WorkspacePodFactory{workspace, params}
}

func (fs *WorkspacePodFactory) DeploymentID() string {
	return fs.workspace.ID
}

func getKubeVolume(params WorkspacePodParameters) []v1.Volume {
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
	return kubeVolume
}

func getKubeVolumeMount(params WorkspacePodParameters) []v1.VolumeMount {
	kubeVolumeMount := []v1.VolumeMount{}
	for _, v := range params.Volumes {
		kubeVolumeMount = append(kubeVolumeMount, v1.VolumeMount{
			Name:      v.VolumeMount.Name,
			SubPath:   v.VolumeMount.SubPath,
			MountPath: v.VolumeMount.MountPath,
		})
	}
	return kubeVolumeMount
}

func (fs *WorkspacePodFactory) NewPod(podName string, params WorkspacePodParameters) v1.Pod {
	kubeVolume := getKubeVolume(params)
	kubeVolumeMount := getKubeVolumeMount(params)

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
					VolumeMounts: kubeVolumeMount,
					Ports: []v1.ContainerPort{
						{
							Name:          "workspace-fs",
							ContainerPort: params.Port,
							Protocol:      v1.ProtocolTCP,
						},
					},
				},
			},
			Volumes: kubeVolume,
		},
	}
}
