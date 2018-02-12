package workspacefs

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestingNewFSPod(t *testing.T) {
	image := "gcr.io/linker-aurora/workspace:develop"
	vName := "testVolume"
	mountPath := "/workspace"

	ws := entity.Workspace{
		ID:   bson.NewObjectId(),
		Name: "testing workspace",
		Type: "general",
		MainVolume: entity.PersistentVolumeClaim{
			Name: vName,
		},
	}

	wsPodParameter := WorkspacePodParameters{
		Port:  WorkspaceContainerPort,
		Image: image,
		Volumes: []container.Volume{
			{
				ClaimName: vName,
				VolumeMount: container.VolumeMount{
					Name:      vName,
					MountPath: mountPath,
				},
			},
		},
	}

	podFactory := WorkspacePodFactory{&ws, wsPodParameter}

	pod := podFactory.NewPod(ws.DeploymentID(), map[string]string{})

	assert.Equal(t, pod.Spec.Containers[0].Image, image)
	assert.Equal(t, pod.Spec.Containers[0].Name, ws.DeploymentID())
	assert.Equal(t, pod.Spec.Containers[0].ImagePullPolicy, "IfNotPresent")
	assert.Equal(t, pod.Spec.Containers[0].Ports[0].ContainerPort, WorkspaceContainerPort)
	assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[0].Name, vName)
	assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[0].MountPath, mountPath)
	assert.Equal(t, pod.Spec.Volumes[0].Name, vName)
	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName, vName)
	assert.Equal(t, pod.Spec.RestartPolicy, "Always")
}
