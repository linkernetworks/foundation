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

func NewFromConfig(cf *MongoConfig) *Service {
	return New(cf.Url)
}

func (s *Service) NewSession() *Session {
	return &Session{s.globalSession.Copy()}
}

func (s *Service) NewClient() (*mgo.Session, error) {
	return mgo.Dial(s.Url)
}
