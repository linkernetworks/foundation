package appspawner

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/apps"
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/environment"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podproxy"
	"bitbucket.org/linkernetworks/aurora/src/kubernetes/pod/podtracker"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"bitbucket.org/linkernetworks/aurora/src/workspace"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestAppSpawnerService(t *testing.T) {
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

	spawner := New(cf, clientset, redisService, mongoService)

	userId := bson.NewObjectId()

	ws := entity.Workspace{
		ID:          bson.NewObjectId(),
		Name:        "testing workspace",
		Type:        "general",
		Owner:       userId,
		Environment: &environment.Tensorflow13Environment,
	}

	session := mongoService.NewSession()
	defer session.Close()

	err = session.C(entity.WorkspaceCollectionName).Insert(ws)
	assert.NoError(t, err)
	defer session.C(entity.WorkspaceCollectionName).Remove(bson.M{"_id": ws.ID})

	// ensure that the primary volume is created
	err = workspace.PrepareVolume(session, &ws, kubernetesService)
	assert.NoError(t, err)
	assert.NotNil(t, ws.PrimaryVolume)

	app := ws.FindApp("linkernetworks/tensorflow-notebook:1.3")
	assert.NotNil(t, app)

	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: &ws}
	assert.Equal(t, "webapp-"+ws.ID.Hex()+"-"+app.ID, wsApp.PodName())

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

	t.Logf("Starting webapp: pod=%s", wsApp.PodName())
	_, err = spawner.Start(&ws, app)
	assert.NoError(t, err)

	// allocattte anew podtracker to track the pod is running
	tracker := podtracker.New(clientset, kubernetesService.Config.Namespace, wsApp.PodName())
	tracker.WaitForPhase(v1.PodPhase("Running"))

	t.Logf("Syncing webapp document: pod=%s", wsApp.PodName())
	err = spawner.AddressUpdater.Sync(wsApp)
	assert.NoError(t, err)

	var conn = redisService.GetConnection()
	addr, err := conn.GetString(podproxy.DefaultPrefix + wsApp.PodName() + ":address")
	assert.NoError(t, err)
	assert.True(t, len(addr) > 0)
	t.Logf("pod address: %s", addr)

	t.Logf("Stoping webapp document: pod=%s", wsApp.PodName())
	_, err = spawner.Stop(&ws, app)
	assert.NoError(t, err)

	err = spawner.AddressUpdater.Sync(wsApp)
	assert.NoError(t, err)

	err = spawner.AddressUpdater.Reset(wsApp)
	assert.NoError(t, err)

}

func TestAppsIsRunningSuccess(t *testing.T) {
	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)

	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)

	spawner := New(cf, clientset, redisService, mongoService)

	//Create a workspaceApp
	//Create a k8s Pod
	//Watchout
	userId := bson.NewObjectId()
	ws := entity.Workspace{
		ID:          bson.NewObjectId(),
		Name:        "testing fileserver",
		Type:        "general",
		Owner:       userId,
		Environment: nil,
	}

	app := &apps.FileServerApp
	assert.NotNil(t, app)

	wsApp := &entity.WorkspaceApp{ContainerApp: app, Workspace: &ws}
	pod, err := spawner.NewPod(wsApp)
	assert.Equal(t, "fileserver-"+ws.ID.Hex()+"-"+app.ID, wsApp.PodName())

	_, err = clientset.CoreV1().Pods("default").Create(pod)
	assert.NoError(t, err)
	defer clientset.CoreV1().Pods("default").Delete(wsApp.PodName(), nil)
	err = spawner.checkAppIsRunning(wsApp, 5)
	assert.NoError(t, err)
}
