package notebookspawner

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"bitbucket.org/linkernetworks/aurora/src/workspace/volumemanager"

	v1 "k8s.io/api/core/v1"

	// "bitbucket.org/linkernetworks/aurora/src/service/notebookspawner/notebook"
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

	var notebookImage = "jupyter/minimal-notebook"
	var err error

	//Get mongo service
	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)

	clientset, err := kubernetesService.NewClientset()
	assert.NoError(t, err)

	spawner := New(cf, mongoService.NewSession(), clientset, redisService)

	// proxyURL := "/v1/notebooks/proxy/"
	session := mongoService.NewSession()
	defer session.Close()

	userId := bson.NewObjectId()

	workspace := entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: userId,
	}

	err = session.C(entity.WorkspaceCollectionName).Insert(workspace)
	assert.NoError(t, err)
	defer session.C(entity.WorkspaceCollectionName).Remove(bson.M{"_id": workspace.ID})

	// ensure that the primary volume is created
	vm := volumemanager.New(clientset, session, "default")
	err = vm.CreatePrimaryVolume(&workspace)
	assert.NoError(t, err)
	assert.NotNil(t, workspace.PrimaryVolume)

	notebookID := bson.NewObjectId()
	notebook := entity.Notebook{
		ID:          notebookID,
		Image:       notebookImage,
		WorkspaceID: workspace.ID,
		Url:         cf.Jupyter.BaseURL + "/" + notebookID.Hex(),
		CreatedBy:   userId,
	}
	err = session.C(entity.NotebookCollectionName).Insert(notebook)
	assert.NoError(t, err)
	defer session.C(entity.NotebookCollectionName).Remove(bson.M{"_id": notebook.ID})

	tracker, err := spawner.Start(&notebook)
	assert.NoError(t, err)
	tracker.WaitFor(v1.PodPhase("Running"))

	err = spawner.Updater.Sync(&notebook)
	assert.NoError(t, err)

	_, err = spawner.Stop(&notebook)
	assert.NoError(t, err)

	err = spawner.Updater.Sync(&notebook)
	assert.NoError(t, err)

}
