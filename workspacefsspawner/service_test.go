package workspacefsspawner

import (
	"os"
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/apps"
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
	clientset, err := kubernetesService.NewClientset()
	assert.NoError(suite.T(), err)
	suite.WsService = New(cf, mongoService, clientset, redisService)

	suite.Session = mongoService.NewSession()
}

func (suite *WorkspaceServiceSuite) TestCRUD() {
	vName := "test-pvc-" + bson.NewObjectId().Hex()

	//Setup the workspace
	ws := &entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: bson.NewObjectId(),
		PrimaryVolume: &container.Volume{
			ClaimName: vName,
			VolumeMount: container.VolumeMount{
				Name:      vName,
				MountPath: "fake",
			},
		},
	}

	wsApp := &entity.WorkspaceApp{Workspace: ws, ContainerApp: &apps.FileServerApp}

	err := suite.Session.C(entity.WorkspaceCollectionName).Insert(ws)
	defer suite.Session.C(entity.WorkspaceCollectionName).Remove((bson.M{"_id": ws.ID}))
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.WakeUp(wsApp)
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.getPod(wsApp)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), kerrors.IsNotFound(err))

	_, err = suite.WsService.Stop(wsApp)
	assert.NoError(suite.T(), err)
}

func (suite *WorkspaceServiceSuite) TestRestart() {
	vName := bson.NewObjectId().Hex()
	//Setup the workspace
	ws := &entity.Workspace{
		ID:    bson.NewObjectId(),
		Name:  "testing workspace",
		Type:  "general",
		Owner: bson.NewObjectId(),
		PrimaryVolume: &container.Volume{
			ClaimName: vName,
			VolumeMount: container.VolumeMount{
				Name:      vName,
				MountPath: "fake",
			},
		},
	}

	err := suite.Session.C(entity.WorkspaceCollectionName).Insert(ws)
	defer suite.Session.C(entity.WorkspaceCollectionName).Remove((bson.M{"_id": ws.ID}))
	assert.NoError(suite.T(), err)

	wsApp := &entity.WorkspaceApp{Workspace: ws, ContainerApp: &apps.FileServerApp}

	_, err = suite.WsService.WakeUp(wsApp)
	assert.NoError(suite.T(), err)

	ws.SecondaryVolumes = []container.Volume{
		{
			ClaimName: "testname",
			VolumeMount: container.VolumeMount{
				Name:      "testname2",
				MountPath: "randompath",
			},
		},
	}

	_, err = suite.WsService.getPod(wsApp)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), kerrors.IsNotFound(err))

	_, err = suite.WsService.Restart(wsApp)
	assert.NoError(suite.T(), err)

	_, err = suite.WsService.Stop(wsApp)
	assert.NoError(suite.T(), err)
}

func (suite *WorkspaceServiceSuite) TestCheckAvailability() {
	id := bson.NewObjectId().Hex()

	err := suite.WsService.CheckAvailability(id, nil, 15)
	assert.NoError(suite.T(), err)

	err = suite.WsService.CheckAvailability(id, &container.Volume{
		ClaimName: "nonexistent",
		VolumeMount: container.VolumeMount{
			Name:      "nonexistent",
			MountPath: "Fake",
		},
	}, 10)
	assert.Error(suite.T(), err)
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
