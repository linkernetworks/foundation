package notify

// import (
// 	"bitbucket.org/linkernetworks/aurora/server"
// 	"bitbucket.org/linkernetworks/aurora/src/config"
// 	"bitbucket.org/linkernetworks/aurora/src/entity"
// 	"bitbucket.org/linkernetworks/aurora/src/utils/timeutils"
//
// 	"github.com/stretchr/testify/assert"
//
// 	"gopkg.in/mgo.v2/bson"
//
// 	"log"
// 	"testing"
// )
//
// func TestNewFailedNotification(t *testing.T) {
// 	var err error
// 	cf := config.MustRead("../../../config/testing.json")
// 	sp := server.NewServiceProviderFromConfig(cf)
// 	assert.NotNil(t, sp)
//
// 	session := sp.Mongo.NewSession()
// 	defer session.Close()
//
// 	userId := bson.NewObjectId()
//
// 	// new a user
// 	newUser := &entity.User{
// 		ID:        userId,
// 		Email:     "develop@linkernetworks.com",
// 		FirstName: "john",
// 		LastName:  "lin",
// 		Cellphone: "0987654556",
// 		Roles:     nil,
// 		Verified:  false,
// 		Revoked:   false,
// 	}
// 	err = session.C(entity.UserCollectionName).Insert(newUser)
// 	assert.NoError(t, err)
//
// 	// a job related to the new user
// 	job := &entity.Job{
// 		ID:          bson.ObjectId("123456789012"),
// 		Name:        "",
// 		Status:      "",
// 		Phase:       "",
// 		WorkspaceID: bson.ObjectId("123456789012"),
// 		Retry:       0,
// 		Priority:    0.0,
// 		CurrentStep: 0,
// 		StartedAt:   timeutils.Now(),
// 		CreatedBy:   userId,
// 		CreatedAt:   timeutils.Now(),
// 	}
//
// 	n, err := NewFailedJobNotification(sp, job)
// 	assert.NotNil(t, n)
// 	assert.NoError(t, err)
// 	log.Printf("%+v", n)
//
// 	err = session.Remove(entity.UserCollectionName, "_id", userId)
// 	assert.NoError(t, err)
// }
//
// func TestNewSucceedNotification(t *testing.T) {
// 	var err error
// 	cf := config.MustRead("../../../config/testing.json")
// 	sp := server.NewServiceProviderFromConfig(cf)
// 	assert.NotNil(t, sp)
//
// 	session := sp.Mongo.NewSession()
// 	defer session.Close()
//
// 	userId := bson.NewObjectId()
//
// 	// new a user
// 	newUser := &entity.User{
// 		ID:        userId,
// 		Email:     "develop@linkernetworks.com",
// 		FirstName: "john",
// 		LastName:  "lin",
// 		Cellphone: "0987654556",
// 		Roles:     nil,
// 		Verified:  false,
// 		Revoked:   false,
// 	}
// 	err = session.C(entity.UserCollectionName).Insert(newUser)
// 	assert.NoError(t, err)
//
// 	// a job related to the new user
// 	job := &entity.Job{
// 		ID:          bson.ObjectId("123456789012"),
// 		Name:        "",
// 		Status:      "",
// 		Phase:       "",
// 		WorkspaceID: bson.ObjectId("123456789012"),
// 		Retry:       0,
// 		Priority:    0.0,
// 		CurrentStep: 0,
// 		StartedAt:   timeutils.Now(),
// 		CreatedBy:   userId,
// 		CreatedAt:   timeutils.Now(),
// 	}
//
// 	n, err := NewSucceedJobNotification(sp, job)
// 	assert.NotNil(t, n)
// 	assert.NoError(t, err)
// 	log.Printf("%+v", n)
//
// 	err = session.Remove(entity.UserCollectionName, "_id", userId)
// 	assert.NoError(t, err)
// }
//
// func TestNewCancledNotification(t *testing.T) {
// 	var err error
// 	cf := config.MustRead("../../../config/testing.json")
// 	sp := server.NewServiceProviderFromConfig(cf)
// 	assert.NotNil(t, sp)
//
// 	session := sp.Mongo.NewSession()
// 	defer session.Close()
//
// 	userId := bson.NewObjectId()
//
// 	// new a user
// 	newUser := &entity.User{
// 		ID:        userId,
// 		Email:     "develop@linkernetworks.com",
// 		FirstName: "john",
// 		LastName:  "lin",
// 		Cellphone: "0987654556",
// 		Roles:     nil,
// 		Verified:  false,
// 		Revoked:   false,
// 	}
// 	err = session.C(entity.UserCollectionName).Insert(newUser)
// 	assert.NoError(t, err)
//
// 	// a job related to the new user
// 	job := &entity.Job{
// 		ID:          bson.ObjectId("123456789012"),
// 		Name:        "",
// 		Status:      "",
// 		Phase:       "",
// 		WorkspaceID: bson.ObjectId("123456789012"),
// 		Retry:       0,
// 		Priority:    0.0,
// 		CurrentStep: 0,
// 		StartedAt:   timeutils.Now(),
// 		CreatedBy:   userId,
// 		CreatedAt:   timeutils.Now(),
// 	}
//
// 	n, err := NewCancledJobNotification(sp, job)
// 	assert.NotNil(t, n)
// 	assert.NoError(t, err)
// 	log.Printf("%+v", n)
//
// 	err = session.Remove(entity.UserCollectionName, "_id", userId)
// 	assert.NoError(t, err)
// }
//
// func TestNewStartedNotification(t *testing.T) {
// 	var err error
// 	cf := config.MustRead("../../../config/testing.json")
// 	sp := server.NewServiceProviderFromConfig(cf)
// 	assert.NotNil(t, sp)
//
// 	session := sp.Mongo.NewSession()
// 	defer session.Close()
//
// 	userId := bson.NewObjectId()
//
// 	// new a user
// 	newUser := &entity.User{
// 		ID:        userId,
// 		Email:     "develop@linkernetworks.com",
// 		FirstName: "john",
// 		LastName:  "lin",
// 		Cellphone: "0987654556",
// 		Roles:     nil,
// 		Verified:  false,
// 		Revoked:   false,
// 	}
// 	err = session.C(entity.UserCollectionName).Insert(newUser)
// 	assert.NoError(t, err)
//
// 	// a job related to the new user
// 	job := &entity.Job{
// 		ID:          bson.ObjectId("123456789012"),
// 		Name:        "",
// 		Status:      "",
// 		Phase:       "",
// 		WorkspaceID: bson.ObjectId("123456789012"),
// 		Retry:       0,
// 		Priority:    0.0,
// 		CurrentStep: 0,
// 		StartedAt:   timeutils.Now(),
// 		CreatedBy:   userId,
// 		CreatedAt:   timeutils.Now(),
// 	}
//
// 	n, err := NewStartedJobNotification(sp, job)
// 	assert.NotNil(t, n)
// 	assert.NoError(t, err)
// 	log.Printf("%+v", n)
//
// 	err = session.Remove(entity.UserCollectionName, "_id", userId)
// 	assert.NoError(t, err)
// }
