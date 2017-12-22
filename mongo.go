package config

type MongoConfig struct {
	Url       string       `json:"url"`
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
