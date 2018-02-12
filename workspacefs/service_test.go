package workspacefs

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"
	"bitbucket.org/linkernetworks/aurora/src/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

type WorkspaceServiceSuite struct {
	suite.Suite
	WsService *WorkspaceFileServerSpawner
	Session   *mongo.Session
}

func (suite *WorkspaceServiceSuite) SetupTest() {

	//Get mongo service
	cf := config.MustRead(testingConfigPath)

	kubernetesService := kubernetes.NewFromConfig(cf.Kubernetes)
	mongoService := mongo.New(cf.Mongo.Url)
	redisService := redis.New(cf.Redis)
	clientset, err := kubernetesService.CreateClientset()
	assert.NoError(suite.T(), err)
	suite.WsService = New(cf, mongoService, clientset, redisService)

	suite.Session = mongoService.NewSession()
}

func (suite *WorkspaceServiceSuite) TestCRUD() {
	vName := bson.NewObjectId().Hex()
	//Setup the workspace
	ws := &entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: bson.NewObjectId(),
		MainVolume: entity.PersistentVolumeClaimParameter{
			Name: vName,
		},
	}

	err := suite.Session.C(entity.WorkspaceCollectionName).Insert(ws)
	defer suite.Session.C(entity.WorkspaceCollectionName).Remove((bson.M{"_id": ws.ID}))
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.WakeUp(ws)
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.getPod(ws)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), kerrors.IsNotFound(err))

	_, err = suite.WsService.Delete(ws)
	assert.NoError(suite.T(), err)
}

func (suite *WorkspaceServiceSuite) TestGetVolume() {
	vName := bson.NewObjectId().Hex()
	ws := &entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: bson.NewObjectId(),
		MainVolume: entity.PersistentVolumeClaimParameter{
			Name: vName,
		},
	}

	volumes, err := suite.WsService.GetKubeVolume(ws)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), volumes[0].ClaimName, vName)
	assert.Equal(suite.T(), volumes[0].VolumeMount.Name, vName)
	assert.Equal(suite.T(), volumes[0].VolumeMount.MountPath, WorkspaceMainVolumeMountPoint)
}

func (suite *WorkspaceServiceSuite) TestRestart() {
	vName := bson.NewObjectId().Hex()
	//Setup the workspace
	ws := &entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: bson.NewObjectId(),
		MainVolume: entity.PersistentVolumeClaimParameter{
			Name: vName,
		},
	}

	err := suite.Session.C(entity.WorkspaceCollectionName).Insert(ws)
	defer suite.Session.C(entity.WorkspaceCollectionName).Remove((bson.M{"_id": ws.ID}))
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.WakeUp(ws)
	assert.NoError(suite.T(), err)

	ws.SubVolumes = []container.Volume{
		{
			ClaimName: "testname",
			VolumeMount: container.VolumeMount{
				Name:      "testname2",
				MountPath: "randompath",
			},
		},
	}

	_, err = suite.WsService.getPod(ws)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), kerrors.IsNotFound(err))

	_, err = suite.WsService.Restart(ws)
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.Delete(ws)
	assert.NoError(suite.T(), err)
}

func TestWorkspaceServiceSuite(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_K8S"); !defined {
		t.SkipNow()
		return
	}
	suite.Run(t, new(WorkspaceServiceSuite))
}

func (suite *WorkspaceServiceSuite) TearDownTest() {
	suite.Session.Close()
}
