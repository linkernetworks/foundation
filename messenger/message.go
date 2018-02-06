package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	_ "bitbucket.org/linkernetworks/aurora/src/logger"
	_ "gopkg.in/mgo.v2/bson"
)

func NewEmail(title, content, from string, to ...string) entity.Notification {
	// sender := from
	sender := "noreply@linkernetworks.com"
	return &entity.Email{
		Title:       title,
		Content:     content,
		ToAddress:   to,
		FromAddress: sender,
	}
}

func NewSMS(content, from string, to ...string) entity.Notification {
	// sender := from
	// test credentials a "From" valid number
	// https://www.twilio.com/docs/api/rest/test-credentials
	sender := "+15005550006"
	return &entity.SMS{
		Content:    content,
		ToNumber:   to,
		FromNumber: sender,
	}
}
