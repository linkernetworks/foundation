package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	url := "a.b.c.d"
	inf := "en0"
	cf := MongoConfig{
		Interface: inf,
		Url:       url,
	}

	assert.Equal(t, cf.GetInterface(), inf)
	assert.False(t, cf.Unresolved())
	assert.Nil(t, cf.GetPublic())
}
