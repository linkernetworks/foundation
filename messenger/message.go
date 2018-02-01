package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"gopkg.in/mgo.v2/bson"
)

type Message interface {
	Title() string
	Content() string
	To() string
	From() string
}

type Email struct {
	msvc *mongo.Service

	title   string
	content string
	to      bson.ObjectId
	from    bson.ObjectId
}

func (e *Email) Title() string {
	return e.title
}

func (e *Email) Content() string {
	return e.content
}

func (e *Email) To() string {
	user, _ := FindUserById(e.msvc, e.to)
	return user.Email
}

func (e *Email) From() string {
	user, _ := FindUserById(e.msvc, e.from)
	return user.Email
}

type SMS struct {
	title   string
	content string
	to      bson.ObjectId
	from    bson.ObjectId
}

func (s *SMS) Title() string {
	return s.title
}

func (s *SMS) Content() string {
	return s.content
}

func (s *SMS) To() string {
	// TODO should use object IDs to find email addresses
	return "123456"
}

func (s *SMS) From() string {
	// TODO should use object IDs to find email addresses
	return "123456"
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
