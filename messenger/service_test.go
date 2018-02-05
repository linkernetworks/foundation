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

	title := "Hello from Mailgun"
	content := `This is a long content. Lorem ipsum dolor sit amet, consectetuer adipiscing 
elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis 
dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque 
eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo,  fringilla vel, 
aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet  a, venenatis vitae, 
justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus 
elementum semper nisi. Aenean vulputate eleifend tellus. Aeneanleo ligula, porttitor eu, 
consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, 
tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. 
Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam 
rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet 
adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, 
lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis 
faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed 
fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, 
augue velit cursus nunc,`

	e := NewEmail(title, content, toUser, fromUser)
	assert.NotNil(t, e)

	mg := NewMailgunService(ms)
	err := mg.Send(e)
	assert.NoError(t, err)
}

func TestTwilioNewService(t *testing.T) {
	const testingConfigPath = "../../../config/testing.json"

	cf := config.MustRead(testingConfigPath)

	ms := mongo.New(cf.Mongo.Url)
	assert.NotNil(t, ms)

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
		Email:     "noreply@linkernetworks.com",
		FirstName: "john",
		LastName:  "lin",
		Cellphone: "+19284409015",
		Roles:     nil,
		Verified:  false,
		Revoked:   false,
	}
	content := "Hello from Twillio. This is the test case message for tesing TestTwilioNewService"

	sms := NewSMS(content, toUser, fromUser)
	assert.NotNil(t, sms)

	twlo := NewTwilioService(ms)
	err := twlo.Send(sms)
	assert.NoError(t, err)
}
