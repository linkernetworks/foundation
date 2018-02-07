package fileserver

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestFileServerSpawnerService(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

	var fileserverImage = "gcr.io/linker-aurora/filerserver:develop"
	var err error

	//Get mongo service
	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)
	fs := New(cf, mongoService, kubernetesService, redisService)

	// proxyURL := "/v1/fileservers/proxy/"
	context := mongoService.NewSession()
	defer context.Close()

	userId := bson.NewObjectId()

	workspace := entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: userId,
	}

	err = context.C(entity.WorkspaceCollectionName).Insert(workspace)
	assert.NoError(t, err)
	defer context.C(entity.WorkspaceCollectionName).Remove(bson.M{"_id": workspace.ID})

	fileserverID := bson.NewObjectId()
	fileserver := entity.FileServer{
		ID:          fileserverID,
		Image:       fileserverImage,
		WorkspaceID: workspace.ID,
		Url:         cf.Jupyter.BaseUrl + "/" + fileserverID.Hex(),
		CreatedBy:   userId,
	}
	err = context.C(entity.FileServerCollectionName).Insert(fileserver)
	assert.NoError(t, err)
	defer context.C(entity.FileServerCollectionName).Remove(bson.M{"_id": fileserver.ID})

	_, err = fs.Start(&fileserver)
	assert.NoError(t, err)

	assert.NoError(t, err)
	_, err = fs.Stop(&fileserver)
	assert.NoError(t, err)
}
