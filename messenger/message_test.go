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

func TestFindUserById(t *testing.T) {
	const testingConfigPath = "../../../config/testing.json"
	var err error
	cf := config.MustRead(testingConfigPath)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

	context := ms.NewSession()
	defer context.Close()

	newUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "develop@linkernetworks.com",
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
	assert.Equal(t, "develop@linkernetworks.com", found.Email)

	err = context.DropCollection(entity.UserCollectionName)
	assert.NoError(t, err)
}

func TestNewEmail(t *testing.T) {
	toUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "develop@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	fromUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "noreply@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "0987654556",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	title := "Hello world"
	content := "This is a long content. This is a long content. This is a long content. This is a long content."

	e := NewEmail(title, content, toUser, fromUser)
	assert.NotNil(t, e)
	assert.Equal(t, "Hello world", e.GetTitle())
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", e.GetContent())
	assert.Equal(t, "develop@linkernetworks.com", e.GetReceiverAddress())
	assert.Equal(t, "noreply@linkernetworks.com", e.GetSenderAddress())
}

func TestNewSMS(t *testing.T) {
	toUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "develop@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "+886952301269",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	fromUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "develop@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "+886952301269",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}

	content := "This is a long content. This is a long content. This is a long content. This is a long content."

	sms := NewSMS(content, toUser, fromUser)
	assert.NotNil(t, sms)
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", sms.GetContent())
	assert.Equal(t, "+886952301269", sms.GetReceiverPhoneNumber())
	assert.Equal(t, "+19284409015", sms.GetSenderPhoneNumber())
}
