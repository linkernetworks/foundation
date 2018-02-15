package config

type VenderDatabase struct {
	Influxdb struct {
		Url      string `json:"url" bson:"url"`
		Category struct {
			Air struct {
				Database string `json:"database" bson:"database"`
			} `json:"air" bson:"air"`
			Water struct {
				Database string `json:"database" bson:"database"`
			} `json:"water" bson:"water"`
		} `json:"category" bson:"category"`
	} `json:"influxdb" bson:"influxdb"`
}

type Lassnet struct {
	Air struct {
		BaseUrl string                   `json:"base_url" bson:"base_url"`
		Devices []map[string]interface{} `json:"devices" bson:"devices"`

		History struct {
			Path string `json:"path" bson:"path"`
		} `json:"history" bson:"history"`

		Latest struct {
			Path string `json:"path" bson:"path"`
		} `json:"latest" bson:"latest"`

		Date struct {
			Path string `json:"path" bson:"path"`
		} `json:"date" bson:"date"`
	} `json:"air" bson:"air"`
}

type Quarkioe struct {
	Water struct {
		BaseUrl string                   `json:"base_url" bson:"base_url"`
		Devices []map[string]interface{} `json:"devices" bson:"devices"`

		Auth struct {
			User     string `json:"user" bson:"user"`
			Password string `json:"password" bson:"password"`
		} `json:"Auth" bson:"Auth"`
	} `json:"water" bson:"water"`
}

type VenderDetail struct {
	Lassnet  Lassnet  `json:"lassnet" bson:"lassnet"`
	Quarkioe Quarkioe `json:"quarkioe" bson:"quarkioe"`
}

type VenderConfig struct {
	Database VenderDatabase `json:"database" bson:"database"`
	Vender   VenderDetail   `json:"vender" bson:"vender"`
}
