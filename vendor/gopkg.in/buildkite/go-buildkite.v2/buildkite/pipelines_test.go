package buildkite

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestPipelinesService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/pipelines", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	pipelines, _, err := client.Pipelines.List("my-great-org", nil)
	if err != nil {
		t.Errorf("Pipelines.List returned error: %v", err)
	}

	want := []Pipeline{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(pipelines, want) {
		t.Errorf("Pipelines.List returned %+v, want %+v", pipelines, want)
	}
}
