package mongo

import (
	"gopkg.in/mgo.v2"
)

type MongoService struct {
	Url           string
	globalSession *mgo.Session
}

func New(url string) *MongoService {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return &MongoService{
		Url:           url,
		globalSession: session,
	}
}

func (s *MongoService) NewContext() *Context {
	return &Context{
		Session: s.globalSession.Copy(),
	}
}
