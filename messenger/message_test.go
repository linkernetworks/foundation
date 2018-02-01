package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	_ "log"
	"testing"
)

const (
	testingConfigPath = "../../../config/testing.json"
)

func TestFindUserById(t *testing.T) {
	var err error
	cf := config.MustRead(testingConfigPath)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

	context := ms.NewSession()
	defer context.Close()

	newUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "hello@gmail.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	err = context.C(entity.UserCollectionName).Insert(newUser)
	assert.NoError(t, err)

	result := entity.User{}
	err = context.C(entity.UserCollectionName).Find(bson.M{"first_name": "john"}).One(&result)
	assert.NoError(t, err)

	found, err := FindUserById(ms, result.ID)
	assert.NoError(t, err)
	assert.Equal(t, "hello@gmail.com", found.Email)

	err = context.DropCollection(entity.UserCollectionName)
	assert.NoError(t, err)
}

func TestEmail(t *testing.T) {
	cf := config.MustRead(testingConfigPath)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

	context := ms.NewSession()
	defer context.Close()

	newUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "hello@gmail.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	err := context.C(entity.UserCollectionName).Insert(newUser)
	assert.NoError(t, err)

	result := entity.User{}
	err = context.C(entity.UserCollectionName).Find(bson.M{"first_name": "john"}).One(&result)
	assert.NoError(t, err)

	e := &Email{
		msvc:    ms,
		title:   "Hello world",
		content: "This is a long content. This is a long content. This is a long content. This is a long content.",
		to:      result.ID,
		from:    result.ID,
	}
	assert.NotNil(t, e)
	assert.Equal(t, "Hello world", e.Title())
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", e.Content())
	assert.Equal(t, "hello@gmail.com", e.To())
	assert.Equal(t, "hello@gmail.com", e.From())

	err = context.DropCollection(entity.UserCollectionName)
	assert.NoError(t, err)
}
