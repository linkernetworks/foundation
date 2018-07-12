package session

import (
	"github.com/boj/redistore"
)

var Service *service

type service struct {
	Config *configuration
	Store  *redistore.RediStore
}

type configuration struct {
	Size     int
	Network  string
	Address  string
	Password string
	KeyPair  []byte
	MaxAge   int
}

func NewService(size int, network string, address string, password string, sessionAge int, secret string) error {
	keyPair := []byte(secret)
	config := &configuration{
		Size:     size,
		Network:  network,
		Address:  address,
		Password: password,
		KeyPair:  keyPair,
		MaxAge:   sessionAge,
	}
	store, err := redistore.NewRediStore(size, network, address, password, keyPair)
	if err != nil {
		return err
	}

	Service = &service{
		Config: config,
		Store:  store,
	}
	// session age in second
	Service.Store.SetMaxAge(sessionAge)
	return nil
}
