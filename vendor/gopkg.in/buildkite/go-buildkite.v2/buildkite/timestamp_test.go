package buildkite

import (
	"encoding/json"
	"testing"
	"time"
)

const (
	referenceTimeStr = `"2015-03-04T02:26:54Z"`
)

var (
	referenceTime = time.Date(2015, time.March, 4, 2, 26, 54, 0, time.UTC)
)

func TestTimestamp_Marshal(t *testing.T) {
	testCases := []struct {
		desc    string
		data    Timestamp
		want    string
		wantErr bool
		equal   bool
	}{
		{"Reference", Timestamp{referenceTime}, referenceTimeStr, false, true},
		{"Mismatch", Timestamp{}, referenceTimeStr, false, false},
	}
	for _, tc := range testCases {
		out, err := json.Marshal(tc.data)
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: gotErr=%v, wantErr=%v, err=%v", tc.desc, gotErr, tc.wantErr, err)
		}
		got := string(out)
		equal := got == tc.want
		if (got == tc.want) != tc.equal {
			t.Errorf("%s: got=%s, want=%s, equal=%v, want=%v", tc.desc, got, tc.want, equal, tc.equal)
		}
	}
}

func TestTimestamp_Unmarshal(t *testing.T) {

	testCases := []struct {
		desc    string
		data    string
		want    Timestamp
		wantErr bool
		equal   bool
	}{
		{"Reference", referenceTimeStr, Timestamp{referenceTime}, false, true},
		{"Mismatch", referenceTimeStr, Timestamp{}, false, false},
		{"Invalid", `"asdf"`, Timestamp{referenceTime}, true, false},
	}
	for _, tc := range testCases {
		var got Timestamp
		err := json.Unmarshal([]byte(tc.data), &got)
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: gotErr=%v, wantErr=%v, err=%v", tc.desc, gotErr, tc.wantErr, err)
			continue
		}
		equal := got.Equal(tc.want)
		if equal != tc.equal {
			t.Errorf("%s: got=%#v, want=%#v, equal=%v, want=%v", tc.desc, got, tc.want, equal, tc.equal)
		}
	}
}
