package mongo

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestNewFromConfig(t *testing.T) {
	cf := MongoConfig{
		Url: "mongodb://localhost:27017/aurora",
	}
	m := NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)
}

func TestNewSession(t *testing.T) {
	cf := MongoConfig{
		Url: "mongodb://localhost:27017/aurora",
	}
	m := NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)

	s := m.NewSession()
	assert.NotNil(t, s)
}

func TestSessionFind(t *testing.T) {
	cf := MongoConfig{
		Url: "mongodb://localhost:27017/aurora",
	}
	m := NewFromConfig(cf.Mongo)
	s := m.NewSession()
	var records []map[string]interface{}
	err := s.FindAll("notebooks", bson.M{}, &records)
	assert.NoError(t, err)
}
