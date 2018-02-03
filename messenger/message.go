package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"
)

type Message interface {
	GetTitle() string
	GetContent() string
	GetTo() bson.ObjectId
	GetFrom() bson.ObjectId
	GetSenderAddress() string
	GetReceiverAddress() string
	GetSenderPhoneNumber() string
	GetReceiverPhoneNumber() string
}

type Notification struct {
	Content string
	To      bson.ObjectId
	From    bson.ObjectId
}

func (n *Notification) GetContent() string {
	return n.Content
}

func (n *Notification) GetTo() bson.ObjectId {
	return n.To
}

func (n *Notification) GetFrom() bson.ObjectId {
	return n.From
}

type Email struct {
	Notification
	Title       string
	ToAddress   string
	FromAddress string
}

func NewEmail(ms *mongo.Service, title, content string, to, from bson.ObjectId) *Email {
	fromAddress, _ := FindUserById(ms, from)
	sender := fromAddress.Email

	toAddress, _ := FindUserById(ms, to)
	receiver := toAddress.Email

	return &Email{
		Notification: Notification{
			Content: content,
			To:      to,
			From:    from,
		},
		Title:       title,
		ToAddress:   receiver,
		FromAddress: sender,
	}
}

func (e *Email) GetTitle() string {
	return e.Title
}

func (e *Email) GetSenderAddress() string {
	return e.FromAddress
}

func (e *Email) GetReceiverAddress() string {
	return e.ToAddress
}

type SMS struct {
	Notification
	ToNumber   string
	FromNumber string
}

func NewSMS(ms *mongo.Service, content string, to, from bson.ObjectId) *SMS {
	// fromNumber, _ := FindUserById(ms, from)
	// FIXME the trial account can not use custom phone number
	sender := "+19284409015"

	toNumber, _ := FindUserById(ms, to)
	receiver := toNumber.Cellphone
	return &SMS{
		Notification: Notification{
			Content: content,
			To:      to,
			From:    from,
		},
		ToNumber:   receiver,
		FromNumber: sender,
	}
}

func (s *SMS) GetSenderPhoneNumber() string {
	return s.FromNumber
}

func (s *SMS) GetReceiverPhoneNumber() string {
	return s.ToNumber
}

func FindUserById(ms *mongo.Service, uid bson.ObjectId) (*entity.User, error) {
	context := ms.NewSession()
	defer context.Close()
	user := &entity.User{}
	err := context.C(entity.UserCollectionName).FindId(uid).One(user)
	if err != nil {
		logger.Errorf("Find user fail: %s", err.Error())
		return nil, err
	}
	return user, nil
}
