package messenger

import (
	// "bitbucket.org/linkernetworks/aurora/src/entity"
	// "gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
	// "log"
	"testing"
)

func TestnewEmail(t *testing.T) {
	fromUserEmail := "noreply@linkernetworks.com"
	toUserEmail := "develop@linkernetworks.com"

	title := "Hello world"
	content := "This is a long content. This is a long content. This is a long content. This is a long content."

	e := NewEmail(title, content, fromUserEmail, toUserEmail)
	assert.NotNil(t, e)
	assert.Equal(t, "Hello world", e.GetTitle())
	// log.Printf("%+v", e)
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", e.GetContent())
	assert.NotEmpty(t, e.GetTo())
	assert.Equal(t, "noreply@linkernetworks.com", e.GetFrom())
}

func TestnewSMS(t *testing.T) {
	fromUserSMS := "+15005550006"
	toUserSMS := "+886952301269"

	content := "This is a long content. This is a long content. This is a long content. This is a long content."

	sms := NewSMS(content, fromUserSMS, toUserSMS)
	assert.NotNil(t, sms)
	assert.Equal(t, "This is a long content. This is a long content. This is a long content. This is a long content.", sms.GetContent())
	assert.NotEmpty(t, sms.GetTo())
	assert.Equal(t, "+15005550006", sms.GetFrom())
}
