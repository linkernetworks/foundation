package mongo

import (
	"gopkg.in/mgo.v2"
)

type Service struct {
	Url           string
	globalSession *mgo.Session
}

func New(url string) *Service {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return &Service{
		Url:           url,
		globalSession: session,
	}
}

func (s *Service) NewSession() *Session {
	return &Session{
		Session: s.globalSession.Copy(),
	}
}
