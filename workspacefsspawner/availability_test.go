package workspacefsspawner

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/types/container"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "default"
)

func TestMountSuccess(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.CreateClientset()

	id := bson.NewObjectId().Hex()
	volume := []container.Volume{}
	//Deploy a Check POD
	pod := NewAvailablePod(id, volume)
	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	assert.NoError(t, err)
	//Wait the POD
	err = WaitAvailiablePod(clientset, namespace, newPod.ObjectMeta.Name, 10)
	assert.NoError(t, err)

	clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
}

func TestMountFail(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	cf := config.MustRead(testingConfigPath)
	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	clientset, err := kubernetesService.CreateClientset()

	id := bson.NewObjectId().Hex()
	volume := []container.Volume{
		{
			ClaimName: "unexist",
			VolumeMount: container.VolumeMount{
				Name:      "unexist",
				MountPath: "aaa",
			},
		},
	}
	//Deploy a Check POD
	pod := NewAvailablePod(id, volume)
	assert.NotNil(t, pod)

	newPod, err := clientset.CoreV1().Pods(namespace).Create(&pod)
	assert.NoError(t, err)
	//Wait the POD
	err = WaitAvailiablePod(clientset, namespace, newPod.ObjectMeta.Name, 4)

	assert.Error(t, err)
	assert.Equal(t, err, ErrMountUnAvailable)
	clientset.CoreV1().Pods(namespace).Delete(newPod.ObjectMeta.Name, &metav1.DeleteOptions{})
}
