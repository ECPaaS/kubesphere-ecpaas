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
	"strconv"
	"time"
)

// type GPUMetadata struct {
// 	Metric string `json:"metric,omitempty" description:"metric name"`
// 	Type   string `json:"type,omitempty" description:"metric type"`
// 	Help   string `json:"help,omitempty" description:"metric description"`
// }

type GPUMetric struct {
	MetricName    string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum" csv:"metric_name"`
	GPUMetricData `json:"data,omitempty" description:"actual metric result"`
	Error         string `json:"error,omitempty" csv:"-"`
}

type GPUMetricValues []GPUMetricValue

type GPUMetricData struct {
	MetricType      string `json:"resultType,omitempty" description:"result type, one of matrix, vector" csv:"metric_type"`
	GPUMetricValues `json:"result,omitempty" description:"metric data including labels, time series and values" csv:"metric_values"`
}

// The first element is the timestamp, the second is the metric value.
// eg, [1585658599.195, 0.528]
type GPUPoint struct {
	Timestamp float64
	Value     float64
}

type GPUMetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"time series labels"`
	// The type of Point is a float64 array with fixed length of 2.
	// So Point will always be initialized as [0, 0], rather than nil.
	// To allow empty Sample, we should declare Sample to type *Point
	Sample         *GPUPoint        `json:"value,omitempty" description:"time series, values of vector type"`
	Series         []GPUPoint       `json:"values,omitempty" description:"time series, values of matrix type"`
	ExportSample   *GPUExportPoint  `json:"exported_value,omitempty" description:"exported time series, values of vector type"`
	ExportedSeries []GPUExportPoint `json:"exported_values,omitempty" description:"exported time series, values of matrix type"`

	MinValue     string `json:"min_value" description:"minimum value from monitor points"`
	MaxValue     string `json:"max_value" description:"maximum value from monitor points"`
	AvgValue     string `json:"avg_value" description:"average value from monitor points"`
	SumValue     string `json:"sum_value" description:"sum value from monitor points"`
	Fee          string `json:"fee" description:"resource fee"`
	ResourceUnit string `json:"resource_unit"`
	CurrencyUnit string `json:"currency_unit"`
}

func (mv *GPUMetricValue) TransferToExportedMetricValue() {

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

// func (p GPUPoint) Timestamp() float64 {
// 	return p[0]
// }

// func (p GPUPoint) Value() float64 {
// 	return p[1]
// }

func (p GPUPoint) transferToExported() GPUExportPoint {
	return GPUExportPoint{p.Timestamp, p.Value}
}

func (p GPUPoint) Add(other GPUPoint) GPUPoint {
	return GPUPoint{p.Timestamp, p.Value + other.Value}
}

// // MarshalJSON implements json.Marshaler. It will be called when writing JSON to HTTP response
// // Inspired by prometheus/client_golang
// func (p GPUPoint) MarshalJSON() ([]byte, error) {
// 	t, err := jsoniter.Marshal(p.Timestamp)
// 	if err != nil {
// 		return nil, err
// 	}
// 	v, err := jsoniter.Marshal(strconv.FormatFloat(p.Value, 'f', -1, 64))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []byte(fmt.Sprintf("[%s,%s]", t, v)), nil
// }

// // UnmarshalJSON implements json.Unmarshaler. This is for unmarshaling test data.
// func (p *GPUPoint) UnmarshalJSON(b []byte) error {
// 	var v []interface{}
// 	if err := jsoniter.Unmarshal(b, &v); err != nil {
// 		return err
// 	}

// 	if v == nil {
// 		return nil
// 	}

// 	if len(v) != 2 {
// 		return errors.New("unsupported array length")
// 	}

// 	ts, ok := v[0].(float64)
// 	if !ok {
// 		return errors.New("failed to unmarshal [timestamp]")
// 	}
// 	valstr, ok := v[1].(string)
// 	if !ok {
// 		return errors.New("failed to unmarshal [value]")
// 	}
// 	valf, err := strconv.ParseFloat(valstr, 64)
// 	if err != nil {
// 		return err
// 	}

// 	p.Timestamp = ts
// 	p.Value = valf
// 	return nil
// }

type GPUCSVPoint struct {
	MetricName   string `csv:"metric_name"`
	Selector     string `csv:"selector"`
	Time         string `csv:"time"`
	Value        string `csv:"value"`
	ResourceUnit string `csv:"unit"`
}

type GPUExportPoint struct {
	Timestamp float64
	Value     float64
}

func (p GPUExportPoint) TimestampSeconds() string {
	return time.Unix(int64(p.Timestamp), 0).Format("2006-01-02 03:04:05 PM")
}

// func (p GPUExportPoint) Value() float64 {
// 	return p[1]
// }

func (p GPUExportPoint) Format() string {
	return strconv.FormatFloat(p.Timestamp, 'f', -1, 64) + " " + strconv.FormatFloat(p.Value, 'f', -1, 64)
}

func (p GPUExportPoint) TransformToCSVPoint(metricName string, selector string, resourceUnit string) GPUCSVPoint {
	return GPUCSVPoint{
		MetricName:   metricName,
		Selector:     selector,
		Time:         p.TimestampSeconds(),
		Value:        strconv.FormatFloat(p.Value, 'f', -1, 64),
		ResourceUnit: resourceUnit,
	}
}
