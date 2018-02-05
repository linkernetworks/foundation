package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	_ "bitbucket.org/linkernetworks/aurora/src/logger"
	_ "gopkg.in/mgo.v2/bson"
)

type MessageType int

const (
	EmailMessage MessageType = 1 << iota
	SMSMessage
)

func NewMessage(t MessageType, title, content string, to, from *entity.User) entity.Messenger {
	switch t {
	case EmailMessage:
		return newEmail(title, content, to, from)
	case SMSMessage:
		return newSMS(content, to, from)
	default:
		return nil
	}
}

func newEmail(title, content string, to, from *entity.User) entity.Messenger {
	sender := "noreply@linkernetworks.com"
	return &entity.Email{
		Title:       title,
		Content:     content,
		ToAddress:   to.Email,
		FromAddress: sender,
	}
}

func newSMS(content string, to, from *entity.User) entity.Messenger {
	sender := "+19284409015"
	return &entity.SMS{
		Content:    content,
		ToNumber:   to.Cellphone,
		FromNumber: sender,
	}
}
