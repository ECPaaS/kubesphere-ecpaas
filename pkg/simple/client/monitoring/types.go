/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package monitoring

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	MetricTypeMatrix = "matrix"
	MetricTypeVector = "vector"
)

type Metadata struct {
	Metric string `json:"metric,omitempty" description:"metric name"`
	Type   string `json:"type,omitempty" description:"metric type"`
	Help   string `json:"help,omitempty" description:"metric description"`
}

type Metric struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum" csv:"metric_name"`
	MetricData `json:"data,omitempty" description:"actual metric result. If error exsit, this feild is omitted"`
	Error      string `json:"error,omitempty" description:"error reason. If data field has value, this filed is omitted" csv:"-"`
}

type MetricValues []MetricValue

type MetricData struct {
	MetricType   string `json:"resultType,omitempty" description:"result type, one of matrix, vector" csv:"metric_type"`
	MetricValues `json:"result,omitempty" description:"metric data including metric labels, time series values and min/max/sum/avg value" csv:"metric_values"`
}

type DashboardEntity struct {
	GrafanaDashboardUrl     string `json:"grafanaDashboardUrl,omitempty"`
	GrafanaDashboardContent string `json:"grafanaDashboardContent,omitempty"`
	Description             string `json:"description,omitempty"`
	Namespace               string `json:"namespace,omitempty"`
}

// The first element is the timestamp, the second is the metric value.
// eg, [1585658599.195, 0.528]
type Point struct {
	Timestamp float64
	Value     float64
}

type ExportPoint struct {
	TimestampExportPoint float64
	ValueExportPoint     float64
}

type MetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"map string object, time series labels. eg. key1:value1, key2:value2"`
	// The type of Point is a float64 array with fixed length of 2.
	// So Point will always be initialized as [0, 0], rather than nil.
	// To allow empty Sample, we should declare Sample to type *Point
	Sample         *Point        `json:"value,omitempty" description:"If resultType=vector. time series, If resultType=vector and query field opertation is qeury, values is vector type"`
	Series         []Point       `json:"values,omitempty" description:"If resultType=matrix. time series. If resultType=vector and query field opertation is query, values is matrix type"`
	ExportSample   *ExportPoint  `json:"exported_value,omitempty" description:"when query field operation is export. exported time series. If resultType=vector and query field opertation is export, value is vector type"`
	ExportedSeries []ExportPoint `json:"exported_values,omitempty" description:"when query field operation is export. exported time series, If resultType=matrix and query field opertation is export, values is matrix type"`

	MinValue     string `json:"min_value" description:"minimum value from monitor points"`
	MaxValue     string `json:"max_value" description:"maximum value from monitor points"`
	AvgValue     string `json:"avg_value" description:"average value from monitor points"`
	SumValue     string `json:"sum_value" description:"sum value from monitor points"`
	Fee          string `json:"fee" description:"resource used fee"`
	ResourceUnit string `json:"resource_unit" description:"resource unit eg. percentages, cores or bytes"`
	CurrencyUnit string `json:"currency_unit" description:"currency code ref https://www.iban.com/currency-codes"`
}

func (mv *MetricValue) TransferToExportedMetricValue() {

	if mv.Sample != nil {
		sample := mv.Sample.transferToExported()
		mv.ExportSample = &sample
		mv.Sample = nil
	}

	for _, item := range mv.Series {
		mv.ExportedSeries = append(mv.ExportedSeries, item.transferToExported())
	}
	mv.Series = nil

	return
}

// func (p Point) Timestamp() float64 {
// 	return p[0]
// }

// func (p Point) Value() float64 {
// 	return p[1]
// }

func (p Point) transferToExported() ExportPoint {
	return ExportPoint{p.Timestamp, p.Value}
}

func (p Point) Add(other Point) Point {
	return Point{p.Timestamp, p.Value + other.Value}
}

// MarshalJSON implements json.Marshaler. It will be called when writing JSON to HTTP response
// Inspired by prometheus/client_golang
func (p Point) MarshalJSON() ([]byte, error) {
	t, err := jsoniter.Marshal(p.Timestamp)
	if err != nil {
		return nil, err
	}
	v, err := jsoniter.Marshal(strconv.FormatFloat(p.Value, 'f', -1, 64))
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[%s,%s]", t, v)), nil
}

// UnmarshalJSON implements json.Unmarshaler. This is for unmarshaling test data.
func (p *Point) UnmarshalJSON(b []byte) error {
	var v []interface{}
	if err := jsoniter.Unmarshal(b, &v); err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	if len(v) != 2 {
		return errors.New("unsupported array length")
	}

	ts, ok := v[0].(float64)
	if !ok {
		return errors.New("failed to unmarshal [timestamp]")
	}
	valstr, ok := v[1].(string)
	if !ok {
		return errors.New("failed to unmarshal [value]")
	}
	valf, err := strconv.ParseFloat(valstr, 64)
	if err != nil {
		return err
	}

	p.Timestamp = ts
	p.Value = valf
	return nil
}

type CSVPoint struct {
	MetricName   string `csv:"metric_name"`
	Selector     string `csv:"selector"`
	Time         string `csv:"time"`
	Value        string `csv:"value"`
	ResourceUnit string `csv:"unit"`
}

func (p ExportPoint) Timestamp() string {
	return time.Unix(int64(p.TimestampExportPoint), 0).Format("2006-01-02 03:04:05 PM")
}

// func (p ExportPoint) Value() float64 {
// 	return p[1]
// }

func (p ExportPoint) Format() string {
	return p.Timestamp() + " " + strconv.FormatFloat(p.ValueExportPoint, 'f', -1, 64)
}

func (p ExportPoint) TransformToCSVPoint(metricName string, selector string, resourceUnit string) CSVPoint {
	return CSVPoint{
		MetricName:   metricName,
		Selector:     selector,
		Time:         p.Timestamp(),
		Value:        strconv.FormatFloat(p.ValueExportPoint, 'f', -1, 64),
		ResourceUnit: resourceUnit,
	}
}
