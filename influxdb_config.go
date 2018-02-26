package config

type InfluxdbConfig struct {
	Url      string `json:"url" bson:"url"`
	Database string `json:"database" bson:"database"`
	Document string `json:"document" bson:"document"`
}
