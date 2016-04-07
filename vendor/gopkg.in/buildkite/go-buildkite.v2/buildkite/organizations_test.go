package buildkite

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestOrganizationsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	orgs, _, err := client.Organizations.List(nil)
	if err != nil {
		t.Errorf("Organizations.List returned error: %v", err)
	}

	want := []Organization{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(orgs, want) {
		t.Errorf("Organizations.List returned %+v, want %+v", orgs, want)
	}
}

func TestOrganizationsService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/babelstoemp", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":"123"}`)
	})

	org, _, err := client.Organizations.Get("babelstoemp")
	if err != nil {
		t.Errorf("Organizations.Get returned error: %v", err)
	}

	want := &Organization{ID: String("123")}
	if !reflect.DeepEqual(org, want) {
		t.Errorf("Organizations.Get returned %+v, want %+v", org, want)
	}
}
