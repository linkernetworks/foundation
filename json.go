package influxdb

import (
	"errors"
	"github.com/influxdata/influxdb/client/v2"
)

type JsonOutput []Row

type Row map[string]interface{}

func FormatOutputToJson(r *client.Response) (JsonOutput, error) {
	out := JsonOutput{}
	if len(r.Results) > 1 || len(r.Results[0].Series) > 1 {
		return out, errors.New("Format failed. Response has more than one Result or Series")
	}
	series := r.Results[0].Series[0]
	columns := series.Columns

	// row []interfaces
	for _, row := range r.Results[0].Series[0].Values {
		newRow := Row{}
		// map field with columns
		for j, field := range row {
			c := columns[j]
			newRow[c] = field
		}
		out = append(out, newRow)
	}
	return out, nil
}
