package mongo

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"bitbucket.org/linkernetworks/aurora/src/config"
)

func TestNewFromConfig(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	m := NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)
}

func TestNewSession(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	m := NewFromConfig(cf.Mongo)
	assert.NotNil(t, m)

	s := m.NewSession()
	assert.NotNil(t, s)
}

func TestSessionFind(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	m := NewFromConfig(cf.Mongo)
	s := m.NewSession()
	var records []map[string]interface{}
	err := s.FindAll("notebooks", bson.M{}, &records)
	assert.NoError(t, err)
}
