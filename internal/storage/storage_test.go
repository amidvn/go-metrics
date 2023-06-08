package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCounter(t *testing.T) {
	s := New(300, "", false)
	testCases := []struct {
		name        string
		metricsName string
		value       int64
		result      int64
	}{
		{name: "UpdateCounter() Test 1", metricsName: "testCounter1", value: 1, result: 1},
		{name: "UpdateCounter() Test 2", metricsName: "testCounter2", value: 1, result: 1},
		{name: "UpdateCounter() Test 3", metricsName: "testCounter1", value: 1, result: 2},
		{name: "UpdateCounter() Test 4", metricsName: "testCounter1", value: 10000, result: 10002},
		{name: "UpdateCounter() Test 5", metricsName: "testCounter2", value: 0, result: 1},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateCounter(test.metricsName, test.value)
			assert.Equal(t, counter(test.result), s.counterData[test.metricsName])
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	s := New(300, "", false)
	testCases := []struct {
		name        string
		metricsName string
		value       float64
		result      float64
	}{
		{name: "UpdateGauge() Test 1", metricsName: "testGauge1", value: 1, result: 1.0},
		{name: "UpdateGauge() Test 2", metricsName: "testGauge2", value: 1.0, result: 1.0},
		{name: "UpdateGauge() Test 4", metricsName: "testGauge1", value: 10000, result: 10000.0},
		{name: "UpdateGauge() Test 5", metricsName: "testGauge2", value: 0, result: 0.0},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateGauge(test.metricsName, test.value)
			assert.Equal(t, gauge(test.result), s.gaugeData[test.metricsName])
		})
	}
}
