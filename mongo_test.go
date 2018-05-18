package config

import (
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"
	"github.com/stretchr/testify/assert"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

func TestNewFromConfig(t *testing.T) {
	cf := MustRead(testingConfigPath)
	m := mongo.NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)
}

func TestNewSession(t *testing.T) {
	cf := MustRead(testingConfigPath)
	m := mongo.NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)

	s := m.NewSession()
	assert.NotNil(t, s)
}

func TestSessionFind(t *testing.T) {
	cf := MustRead(testingConfigPath)
	m := mongo.NewFromConfig(cf.Mongo)
	s := m.NewSession()
	var records []map[string]interface{}
	err := s.FindAll("notebooks", bson.M{}, &records)
	assert.NoError(t, err)
}
