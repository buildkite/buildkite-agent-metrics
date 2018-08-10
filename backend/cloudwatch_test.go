package backend

import (
	"reflect"
	"testing"
)

func TestParseCloudWatchDimensions(t *testing.T) {
	for _, tc := range []struct {
		s        string
		expected []CloudWatchDimension
	}{
		{"", []CloudWatchDimension{}},
		{" ", []CloudWatchDimension{}},
		{"Key=Value", []CloudWatchDimension{{"Key", "Value"}}},
		{"Key=Value,Another=Value", []CloudWatchDimension{{"Key", "Value"}, {"Another", "Value"}}},
		{"Key=Value, Another=Value", []CloudWatchDimension{{"Key", "Value"}, {"Another", "Value"}}},
		{"Key=Value, Another=Value ", []CloudWatchDimension{{"Key", "Value"}, {"Another", "Value"}}},
	} {
		t.Run(tc.s, func(t *testing.T) {
			d, err := ParseCloudWatchDimensions(tc.s)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(d, tc.expected) {
				t.Errorf("Expected %s to parse to %#v, got %#v", tc.s, tc.expected, d)
			}
		})
	}
}
