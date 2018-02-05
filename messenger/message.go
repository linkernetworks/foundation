package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	_ "bitbucket.org/linkernetworks/aurora/src/logger"
	_ "gopkg.in/mgo.v2/bson"
)

type MessageType int

const (
	EMAIL MessageType = 1 << iota
	SMS
)

func NewNotificationMessage(t MessageType, title, content string, to, from *entity.User) entity.Notification {
	switch t {
	case EMAIL:
		return newEmail(title, content, to, from)
	case SMS:
		return newSMS(content, to, from)
	default:
		return nil
	}
}

func newEmail(title, content string, to, from *entity.User) entity.Notification {
	sender := "noreply@linkernetworks.com"
	return &entity.Email{
		Title:       title,
		Content:     content,
		ToAddress:   to.Email,
		FromAddress: sender,
	}
}

func newSMS(content string, to, from *entity.User) entity.Notification {
	sender := "+19284409015"
	return &entity.SMS{
		Content:    content,
		ToNumber:   to.Cellphone,
		FromNumber: sender,
	}
}
