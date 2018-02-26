package config

type MongoConfig struct {
	Url       string       `json:"url"`
	Database  string       `json:"database" bson:"database"`
	Document  string       `json:"document" bson:"document"`
	Interface string       `json:"interface"`
	Public    *MongoConfig `json:"public"`
}

func (c *MongoConfig) Unresolved() bool {
	return len(c.Url) < 1
}

func (c *MongoConfig) GetInterface() string {
	return c.Interface
}

func (c *MongoConfig) GetPublic() *MongoConfig {
	return c.Public
}
