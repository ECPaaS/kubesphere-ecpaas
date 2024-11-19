// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package monitoring

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	meteringclient "kubesphere.io/kubesphere/pkg/simple/client/metering"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

func TestGetMaxPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 1.0, Value: 2.0},
				{Timestamp: 3.0, Value: 4.0},
			},
			expectedValue: "4.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 2, Value: 1},
				{Timestamp: 4, Value: 3.1},
			},
			expectedValue: "3.1000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 5, Value: 100},
				{Timestamp: 6, Value: 100000.001},
			},
			expectedValue: "100000.0010000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getMaxPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("max point value caculation is wrong.")
			}
		})
	}
}

func TestGetMinPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 1.0, Value: 2.0},
				{Timestamp: 3.0, Value: 4.0},
			},
			expectedValue: "2.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 2, Value: 1},
				{Timestamp: 4, Value: 3.1},
			},
			expectedValue: "1.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 5, Value: 100},
				{Timestamp: 6, Value: 100000.001},
			},
			expectedValue: "100.0000000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getMinPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("min point value caculation is wrong.")
			}
		})
	}
}

func TestGetSumPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 1.0, Value: 2.0},
				{Timestamp: 3.0, Value: 4.0},
			},
			expectedValue: "6.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 2, Value: 1},
				{Timestamp: 4, Value: 3.1},
			},
			expectedValue: "4.1000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 5, Value: 100},
				{Timestamp: 6, Value: 100000.001},
			},
			expectedValue: "100100.0010000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getSumPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("sum point value caculation is wrong.")
			}
		})
	}
}

func TestGetAvgPointValue(t *testing.T) {
	tests := []struct {
		actualPoints  []monitoring.Point
		expectedValue string
	}{
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 1.0, Value: 2.0},
				{Timestamp: 3.0, Value: 4.0},
			},
			expectedValue: "3.0000000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 2, Value: 1},
				{Timestamp: 4, Value: 3.1},
			},
			expectedValue: "2.0500000000",
		},
		{
			actualPoints: []monitoring.Point{
				{Timestamp: 5, Value: 100},
				{Timestamp: 6, Value: 100000.001},
			},
			expectedValue: "50050.0005000000",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			max := getAvgPointValue(tt.actualPoints)
			if max != tt.expectedValue {
				t.Fatal("avg point value caculattion is wrong.")
			}
		})
	}
}

func TestGenerateFloatFormat(t *testing.T) {
	format := generateFloatFormat(10)
	if format != "%.10f" {
		t.Fatalf("get currency float format failed, %s", format)
	}
}

func TestGetResourceUnit(t *testing.T) {

	tests := []struct {
		meterName     string
		expectedValue string
	}{
		{
			meterName:     "no-exist",
			expectedValue: "",
		},
		{
			meterName:     "meter_cluster_cpu_usage",
			expectedValue: "cores",
		},
	}
	for _, tt := range tests {
		if getResourceUnit(tt.meterName) != tt.expectedValue {
			t.Fatal("get resource unit failed")
		}
	}

}

func TestSquashPoints(t *testing.T) {

	tests := []struct {
		input    []monitoring.Point
		factor   int
		expected []monitoring.Point
	}{
		{
			input: []monitoring.Point{
				{Timestamp: 1, Value: 1},
				{Timestamp: 2, Value: 2},
				{Timestamp: 3, Value: 3},
				{Timestamp: 4, Value: 4},
				{Timestamp: 5, Value: 5},
				{Timestamp: 6, Value: 6},
				{Timestamp: 7, Value: 7},
				{Timestamp: 8, Value: 8},
			},
			factor: 1,
			expected: []monitoring.Point{
				{Timestamp: 1, Value: 1},
				{Timestamp: 2, Value: 2},
				{Timestamp: 3, Value: 3},
				{Timestamp: 4, Value: 4},
				{Timestamp: 5, Value: 5},
				{Timestamp: 6, Value: 6},
				{Timestamp: 7, Value: 7},
				{Timestamp: 8, Value: 8},
			},
		},
		{
			input: []monitoring.Point{
				{Timestamp: 1, Value: 1},
				{Timestamp: 2, Value: 2},
				{Timestamp: 3, Value: 3},
				{Timestamp: 4, Value: 4},
				{Timestamp: 5, Value: 5},
				{Timestamp: 6, Value: 6},
				{Timestamp: 7, Value: 7},
				{Timestamp: 8, Value: 8},
			},
			factor: 2,
			expected: []monitoring.Point{
				{Timestamp: 2, Value: 3},
				{Timestamp: 4, Value: 7},
				{Timestamp: 6, Value: 11},
				{Timestamp: 8, Value: 15},
			},
		},
	}

	for _, tt := range tests {
		got := squashPoints(tt.input, tt.factor)
		if diff := cmp.Diff(got, tt.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", tt.expected, diff)
		}
	}
}

func TestGetFeeWithMeterName(t *testing.T) {

	priceInfo := meteringclient.PriceInfo{
		IngressNetworkTrafficPerMegabytesPerHour: 1,
		EgressNetworkTrafficPerMegabytesPerHour:  2,
		CpuPerCorePerHour:                        3,
		MemPerGigabytesPerHour:                   4,
		PvcPerGigabytesPerHour:                   5,
		CurrencyUnit:                             "CNY",
		GpuPerPercentagePerHour:                  6,
		GpuFbPerMegabytesPerHour:                 7,
		GpuPowerPerKilowattPerHour:               8,
		GpuMemPerPercentagePerHour:               9,
	}

	if getFeeWithMeterName("meter_cluster_cpu_usage", "1", priceInfo) != "3.000" {
		t.Error("failed to get fee with meter_cluster_cpu_usage")
		return
	}
	if getFeeWithMeterName("meter_cluster_memory_usage", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_memory_usage")
		return
	}
	if getFeeWithMeterName("meter_cluster_net_bytes_transmitted", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_net_bytes_transmitted")
		return
	}
	if getFeeWithMeterName("meter_cluster_net_bytes_received", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_net_bytes_received")
		return
	}
	if getFeeWithMeterName("meter_cluster_pvc_bytes_total", "0", priceInfo) != "0.000" {
		t.Error("failed to get fee with meter_cluster_pvc_bytes_total")
		return
	}
	if getFeeWithMeterName("meter_workspace_gpu_usage", "1", priceInfo) != "6.000" {
		t.Error("failed to get fee with meter_workspace_gpu_usage")
		return
	}
	if getFeeWithMeterName("meter_workspace_gpu_framebuffer_usage", "1", priceInfo) != "7.000" {
		t.Error("failed to get fee with meter_workspace_gpu_framebuffer_usage")
		return
	}
	if getFeeWithMeterName("meter_workspace_gpu_power_usage", "1000", priceInfo) != "8.000" {
		t.Error("failed to get fee with meter_workspace_gpu_power_usage")
		return
	}
	if getFeeWithMeterName("meter_workspace_gpu_memory_usage", "1", priceInfo) != "9.000" {
		t.Error("failed to get fee with meter_workspace_gpu_memory_usage")
		return
	}
}

func TestUpdateMetricStatData(t *testing.T) {

	priceInfo := meteringclient.PriceInfo{
		IngressNetworkTrafficPerMegabytesPerHour: 1,
		EgressNetworkTrafficPerMegabytesPerHour:  2,
		CpuPerCorePerHour:                        3,
		MemPerGigabytesPerHour:                   4,
		PvcPerGigabytesPerHour:                   5,
		CurrencyUnit:                             "CNY",
		GpuPerPercentagePerHour:                  6,
		GpuFbPerMegabytesPerHour:                 7,
		GpuPowerPerKilowattPerHour:               8,
		GpuMemPerPercentagePerHour:               9,
	}

	tests := []struct {
		metric     monitoring.Metric
		scalingMap map[string]float64
		expected   monitoring.MetricData
	}{
		{
			metric: monitoring.Metric{
				MetricName: "test",
				MetricData: monitoring.MetricData{
					MetricType: monitoring.MetricTypeMatrix,
					MetricValues: []monitoring.MetricValue{
						{
							Metadata: map[string]string{},
							Series: []monitoring.Point{
								{Timestamp: 1, Value: 1},
								{Timestamp: 2, Value: 2},
							},
						},
					},
				},
			},
			scalingMap: map[string]float64{
				"test": 1,
			},
			expected: monitoring.MetricData{
				MetricType: monitoring.MetricTypeMatrix,
				MetricValues: []monitoring.MetricValue{
					{
						Metadata: map[string]string{},
						Series: []monitoring.Point{
							{Timestamp: 1, Value: 1},
							{Timestamp: 2, Value: 2},
						},
						MinValue:     "1.0000000000",
						MaxValue:     "2.0000000000",
						AvgValue:     "1.5000000000",
						SumValue:     "3.0000000000",
						CurrencyUnit: "CNY",
					},
				},
			},
		},
		{
			metric: monitoring.Metric{
				MetricName: "test",
				MetricData: monitoring.MetricData{
					MetricType: monitoring.MetricTypeVector,
					MetricValues: []monitoring.MetricValue{
						{
							Metadata: map[string]string{},
							Sample:   &monitoring.Point{Timestamp: 1, Value: 2},
						},
					},
				},
			},
			scalingMap: nil,
			expected: monitoring.MetricData{
				MetricType: monitoring.MetricTypeVector,
				MetricValues: []monitoring.MetricValue{
					{
						Metadata:     map[string]string{},
						Sample:       &monitoring.Point{Timestamp: 1, Value: 2},
						MinValue:     "2.0000000000",
						MaxValue:     "2.0000000000",
						AvgValue:     "2.0000000000",
						SumValue:     "2.0000000000",
						CurrencyUnit: "CNY",
					},
				},
			},
		},
	}

	for _, test := range tests {
		got := updateMetricStatData(test.metric, test.scalingMap, priceInfo)
		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			return
		}
	}

}
