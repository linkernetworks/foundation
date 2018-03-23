package mongo

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"log"

	"gopkg.in/mgo.v2"
)

type Service struct {
	Url           string
	globalSession *mgo.Session
}

func New(url string) *Service {
	log.Printf("Try to connect to Mongodb at %s", url)
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	return &Service{
		Url:           url,
		globalSession: session,
	}
}

func NewFromConfig(cf *config.MongoConfig) *Service {
	return New(cf.Url)
}

func (s *Service) NewSession() *Session {
	return &Session{s.globalSession.Copy()}
}
