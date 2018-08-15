package migration

import (
	"github.com/linkernetworks/foundation/config"
	"github.com/linkernetworks/foundation/service/mongo"
)

type Container interface {
	GetConfig() config.Config
	GetMongo() *mongo.Service
}
