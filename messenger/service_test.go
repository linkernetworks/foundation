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

func TestMailgunNewService(t *testing.T) {
	const testingConfigPath = "../../../config/testing.json"

	cf := config.MustRead(testingConfigPath)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

	context := ms.NewSession()
	defer context.Close()

	newUser := &entity.User{
		ID:        bson.ObjectId("123456789012"),
		Email:     "linton.tw@gmail.com",
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

	title := "Hello world"
	content := "This is a long content. This is a long content. This is a long content. This is a long content."
	to := result.ID
	from := result.ID

	e := NewEmail(ms, title, content, to, from)
	assert.NotNil(t, e)

	mg := NewMailgunService(ms)
	err = mg.Send(e)
	assert.NoError(t, err)

	err = context.DropCollection(entity.UserCollectionName)
	assert.NoError(t, err)
}
