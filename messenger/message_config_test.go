package messenger

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"github.com/stretchr/testify/assert"
)

func TestGetAllSender(t *testing.T) {
	n := entity.NotificationEventStruct{
		Event: "JobCreated",
	}
	//FIXME: Need load setting from Mongo
	sms := entity.SMSSettings{}

	testReciver := []string{"user1"}
	cfgService := NewConfigService(NewTwilioService(sms))
	cfgService.cfg = entity.NotificationConfig{
		SMS: entity.SenderConfig{
			JobCreated: testReciver,
		},
	}

	senders := cfgService.GetAllSender(&n)
	assert.NotEqual(t, 1, len(senders), "Should only have one sender")
}
