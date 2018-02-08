package fileserver

import (
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestingNewFSPod(t *testing.T) {
	fpd := FileServerPodFactory{}
	image := "gcr.io/linker-aurora/fileserver:develop"

	vName := "testVolume"
	mountPath := "/workspace"
	fpp := FileServerPodParameters{
		Port:  FileServerContainerPort,
		Image: image,
		Volumes: []container.Volume{
			{
				ClaimName: vName,
				Volume: container.VolumeMount{
					Name:      vName,
					MountPath: mountPath,
				},
			},
		},
	}

	name := "TestForFS"
	pod := fpd.NewPod(name, fpp)

	assert.Equal(t, pod.Spec.Containers[0].Image, image)
	assert.Equal(t, pod.Spec.Containers[0].Name, name)
	assert.Equal(t, pod.Spec.Containers[0].ImagePullPolicy, "IfNotPresent")
	assert.Equal(t, pod.Spec.Containers[0].Ports[0].ContainerPort, FileServerContainerPort)
	assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[0].Name, vName)
	assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[0].MountPath, mountPath)
	assert.Equal(t, pod.Spec.Volumes[0].Name, vName)
	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName, vName)
	assert.Equal(t, pod.Spec.RestartPolicy, "Always")
}
