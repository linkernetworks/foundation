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
	Title   string
	Content string
	To      bson.ObjectId
	From    bson.ObjectId
}

func (n *Notification) GetTitle() string {
	return n.Title
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
	toAddress   string
	fromAddress string
}

func NewEmail(ms *mongo.Service, title, content string, to, from bson.ObjectId) *Email {
	fromAddress, _ := FindUserById(ms, from)
	sender := fromAddress.Email

	toAddress, _ := FindUserById(ms, to)
	receiver := toAddress.Email

	return &Email{
		Notification: Notification{
			Title:   title,
			Content: content,
			To:      to,
			From:    from,
		},
		toAddress:   receiver,
		fromAddress: sender,
	}
}

func (e *Email) GetSenderAddress() string {
	return e.toAddress
}

func (e *Email) GetReceiverAddress() string {
	return e.fromAddress
}

type SMS struct {
	Notification
	toNumber   string
	fromNumber string
}

func NewSMS(ms *mongo.Service, title, content string, to, from bson.ObjectId) *SMS {
	fromNumber, _ := FindUserById(ms, from)
	sender := fromNumber.Cellphone

	toNumber, _ := FindUserById(ms, to)
	receiver := toNumber.Cellphone
	return &SMS{
		Notification: Notification{
			Title:   title,
			Content: content,
			To:      to,
			From:    from,
		},
		toNumber:   receiver,
		fromNumber: sender,
	}
}

func (s *SMS) GetSenderPhoneNumber() string {
	return s.toNumber
}

func (s *SMS) GetReceiverPhoneNumber() string {
	return s.fromNumber
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
