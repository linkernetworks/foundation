package fileserver

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

func TestingNewFSPod(t *testing.T) {
	id := bson.NewObjectId()
	fs := entity.FileServer{
		ID:          id,
		Name:        "TestingForFSPOD",
		Description: "",
		WorkspaceID: "",
	}

	fpd := FileServerPodFactory{FileServer: &fs}

	assert.Equal(t, fpd.DeploymentID, id.Hex())

	image := "gcr.io/linker-aurora/fileserver:develop"
	fpp := FileServerPodParameters{
		Port:  FileServerContainerPort,
		Image: image,
	}

	name := "TestForFS"
	pod := fpd.NewPod(name, fpp)
	assert.Equal(t, pod.Spec.Containers[0].Image, image)
	assert.Equal(t, pod.Spec.Containers[0].Name, name)
	assert.Equal(t, pod.Spec.Containers[0].ImagePullPolicy, "IfNotPresent")
	assert.Equal(t, pod.Spec.Containers[0].Ports[0].ContainerPort, FileServerContainerPort)
	assert.Equal(t, pod.Spec.RestartPolicy, "Always")
}
