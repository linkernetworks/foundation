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

func TestFileServerServiceWakeup(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}

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
	fileserver := entity.FileServer{}

	vName := "testmount"
	workspace := entity.Workspace{
		ID:         bson.NewObjectId(),
		Name:       "testing workspace",
		Type:       "general",
		Owner:      userId,
		FileServer: fileserver,
		MainVolume: entity.PersistentVolumeClaim{
			Name: vName,
		},
	}

	err := context.C(entity.WorkspaceCollectionName).Insert(workspace)
	assert.NoError(t, err)
	defer context.C(entity.WorkspaceCollectionName).Remove(bson.M{"_id": workspace.ID})

	err = fs.WakeUp(&workspace)
	assert.NoError(t, err)
	newWP := entity.Workspace{}

	//Check the PodName has been update
	context.C(entity.WorkspaceCollectionName).Find(bson.M{"_id": workspace.ID}).One(&newWP)
	assert.Equal(t, newWP.PodName, WorkspacePodNamePrefix+workspace.ID.Hex())

	err = fs.Delete(&workspace)
	assert.NoError(t, err)
}
