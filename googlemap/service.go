package googlemap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"bitbucket.org/linkernetworks/aurora/src/entity"
)

type GoogleMapService struct {
	Key string
}

func New(key string) *GoogleMapService {
	return &GoogleMapService{
		Key: key,
	}
}

func (s *GoogleMapService) GetPosition(address string) (entity.GeoCode, error) {

	var position entity.GeoCode

	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/geocode/json?key=%s&address=%s",
		s.Key, address)
	rowData, err := http.Get(url)
	defer rowData.Body.Close()

	body, err := ioutil.ReadAll(rowData.Body)
	if err != nil {
		return position, err
	}

	err = json.Unmarshal(body, &position)
	return position, err
}
