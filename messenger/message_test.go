package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	// "log"
	"testing"
)

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

	e := NewMessage(EmailMessage, title, content, toUser, fromUser)
	// e := NewEmail(title, content, toUser, fromUser)
	assert.NotNil(t, e)
	assert.Equal(t, "Hello world", e.GetTitle())
	// log.Printf("%+v", e)
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", e.GetContent())
	assert.Equal(t, "develop@linkernetworks.com", e.GetTo())
	assert.Equal(t, "noreply@linkernetworks.com", e.GetFrom())
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

	sms := NewMessage(SMSMessage, "", content, toUser, fromUser)
	// sms := NewSMS(content, toUser, fromUser)
	assert.NotNil(t, sms)
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", sms.GetContent())
	assert.Equal(t, "+886952301269", sms.GetTo())
	assert.Equal(t, "+19284409015", sms.GetFrom())
}
