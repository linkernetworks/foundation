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
	KeyPairs []byte
}

func NewService(size int, network string, address string, password string, keyPairs []byte) error {
	config := &configuration{
		Size:     size,
		Network:  network,
		Address:  address,
		Password: password,
		KeyPairs: keyPairs,
	}
	store, err := redistore.NewRediStore(size, network, address, password, keyPairs)
	if err != nil {
		return err
	}
	Service = &service{
		Config: config,
		Store:  store,
	}

	// age in second
	// TODO config
	Service.Store.SetMaxAge(3600 * 30)

	return nil
}
