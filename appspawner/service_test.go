package appspawner

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"bitbucket.org/linkernetworks/aurora/src/workspace"

	v1 "k8s.io/api/core/v1"

	// "bitbucket.org/linkernetworks/aurora/src/service/appspawner/notebook"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestNotebookSpawnerService(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	var err error

	//Get mongo service
	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)

	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)

	spawner := New(cf, mongoService, clientset, redisService)

	// proxyURL := "/v1/notebooks/proxy/"
	session := mongoService.NewSession()
	defer session.Close()

	userId := bson.NewObjectId()

	ws := entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: userId,
		EnvironmentSettings: &entity.EnvironmentSettings{
			Type:     "training",
			Training: &environment.TensorflowEnvironment,
		},
	}

	err = session.C(entity.WorkspaceCollectionName).Insert(ws)
	assert.NoError(t, err)
	defer session.C(entity.WorkspaceCollectionName).Remove(bson.M{"_id": ws.ID})

	// ensure that the primary volume is created
	err = workspace.PrepareVolume(session, &ws, kubernetesService)
	assert.NoError(t, err)
	assert.NotNil(t, ws.PrimaryVolume)

	app := ws.FindApp("jupyter/tensorflow-notebook")
	assert.NotNil(t, app)

	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: &ws}
	assert.Equal(t, "notebook-"+ws.ID.Hex(), wsApp.PodName())

	pod, err := spawner.NewPod(wsApp)
	assert.NoError(t, err)
	t.Logf("pod=%s", wsApp.PodName())

	for _, v := range pod.Spec.Volumes {
		t.Logf("Added Volume: %s", v.Name)
	}
	for _, m := range pod.Spec.Containers[0].VolumeMounts {
		if len(m.SubPath) == 0 {
			t.Logf("Added Mount: mount %s at %s", m.Name, m.MountPath)
		} else {
			t.Logf("Added Mount: mount %s from %s at %s", m.Name, m.SubPath, m.MountPath)
		}
	}

	assert.Equal(t, 2, len(pod.Spec.Volumes))
	assert.Equal(t, 2, len(pod.Spec.Containers[0].VolumeMounts))

	t.Logf("Starting notebook: pod=%s", wsApp.PodName())
	_, err = spawner.Start(&ws, app)
	assert.NoError(t, err)

	tracker := podtracker.New(clientset, kubernetesService.Config.Namespace, wsApp.PodName())
	tracker.WaitForPhase(v1.PodPhase("Running"))

	t.Logf("Syncing notebook document: pod=%s", wsApp.PodName())
	err = spawner.Updater.Sync(wsApp)
	assert.NoError(t, err)

	t.Logf("Stoping notebook document: pod=%s", wsApp.PodName())
	_, err = spawner.Stop(&ws, app)
	assert.NoError(t, err)

	err = spawner.Updater.Sync(wsApp)
	assert.NoError(t, err)

	err = spawner.Updater.Reset(wsApp)
	assert.NoError(t, err)
}
