package influxdb

import (
	"github.com/influxdata/influxdb/client/v2"
)

type InfluxdbService struct {
	Url string
}

func (i *InfluxdbService) NewClient() (client.Client, error) {
	return client.NewHTTPClient(client.HTTPConfig{
		Addr: i.Url,
	})
}
