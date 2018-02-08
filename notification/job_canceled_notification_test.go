package notification

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"bitbucket.org/linkernetworks/aurora/src/utils/timeutils"

	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2/bson"

	"log"
	"testing"
)

func TestNewCanceledJobNotification(t *testing.T) {
	var err error
	cf := config.MustRead("../../../config/testing.json")
	mongoService := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, mongoService)
	session := mongoService.NewSession()

	userId := bson.NewObjectId()

	// new a user
	newUser := &entity.User{
		ID:        userId,
		Email:     "develop@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}
	err = session.C(entity.UserCollectionName).Insert(newUser)
	assert.NoError(t, err)

	// a job related to the new user
	job := &entity.Job{
		ID:          bson.ObjectId("123456789012"),
		Name:        "",
		Status:      "",
		Phase:       "",
		WorkspaceID: bson.ObjectId("123456789012"),
		Retry:       0,
		Priority:    0.0,
		CurrentStep: 0,
		StartedAt:   timeutils.Now(),
		CreatedBy:   userId,
		CreatedAt:   timeutils.Now(),
	}

	n, err := NewCanceledJobNotification(session, job)
	assert.NotNil(t, n)
	assert.NoError(t, err)
	log.Printf("%+v", n)

	err = session.Remove(entity.UserCollectionName, "_id", userId)
	assert.NoError(t, err)
}

func TestCanceledRenderContent(t *testing.T) {
	var err error
	cf := config.MustRead("../../../config/testing.json")
	mongoService := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, mongoService)
	session := mongoService.NewSession()

	userId := bson.NewObjectId()

	// new a user
	newUser := &entity.User{
		ID:        userId,
		Email:     "develop@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}
	err = session.C(entity.UserCollectionName).Insert(newUser)
	assert.NoError(t, err)

	// a job related to the new user
	job := &entity.Job{
		ID:          bson.ObjectId("123456789012"),
		Name:        "",
		Status:      "",
		Phase:       "",
		WorkspaceID: bson.ObjectId("123456789012"),
		Retry:       0,
		Priority:    0.0,
		CurrentStep: 0,
		StartedAt:   timeutils.Now(),
		CreatedBy:   userId,
		CreatedAt:   timeutils.Now(),
	}

	n, err := NewCanceledJobNotification(session, job)
	assert.NotNil(t, n)
	assert.NoError(t, err)
	log.Printf("%+v", n)

	content, err := n.RenderContent()
	assert.Contains(t, content, "canceled")
	assert.NoError(t, err)

	err = session.Remove(entity.UserCollectionName, "_id", userId)
	assert.NoError(t, err)
}
