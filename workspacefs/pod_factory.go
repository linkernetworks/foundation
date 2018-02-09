package workspacefs

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const WorkspaceContainerPort = 33333
const WorkspaceImage = "asia.gcr.io/linker-aurora/file-server"
const WorkspaceFSPortName = "workspace-fs"
const WorkspaceMainVolumeMountPoint = "/workspace"

type WorkspacePodParameters struct {
	// Workspace parameters
	Port    int32
	Image   string
	Volumes []container.Volume
}

type WorkspacePodFactory struct {
	workspace *entity.Workspace
	params    WorkspacePodParameters
}

func NewWorkspacePodFactory(workspace *entity.Workspace, params WorkspacePodParameters) *WorkspacePodFactory {
	return &WorkspacePodFactory{workspace, params}
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

func (ws *WorkspacePodFactory) NewPod(podName string, labels map[string]string) v1.Pod {
	params := ws.params
	kubeVolume := getKubeVolume(params)
	kubeVolumeMount := getKubeVolumeMount(params)

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: labels,
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
							Name:          WorkspaceFSPortName,
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
