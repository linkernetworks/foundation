package config

type InfluxdbConfig struct {
	Url      string `json:"url" bson:"url"`
	Database string `json:"database" bson:"database"`
	Series   string `json:"series" bson:"series"`
}
