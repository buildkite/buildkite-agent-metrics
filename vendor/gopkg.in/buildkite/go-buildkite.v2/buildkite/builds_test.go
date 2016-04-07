package buildkite

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestBuildsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	builds, _, err := client.Builds.List(nil)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/pipelines/sup-keith/builds/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":"123"}`)
	})

	build, _, err := client.Builds.Get("my-great-org", "sup-keith", "123")
	if err != nil {
		t.Errorf("Builds.Get returned error: %v", err)
	}

	want := &Build{ID: String("123")}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Builds.Get returned %+v, want %+v", build, want)
	}
}

func TestBuildsService_List_by_status(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"state[]": "running",
			"page":    "2",
		})
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	opt := &BuildsListOptions{
		State:       []string{"running"},
		ListOptions: ListOptions{Page: 2},
	}
	builds, _, err := client.Builds.List(opt)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_List_by_multiple_status(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValuesList(t, r, valuesList{
			{"state[]", "running"},
			{"state[]", "scheduled"},
			{"page", "2"},
		})
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	opt := &BuildsListOptions{
		State:       []string{"running", "scheduled"},
		ListOptions: ListOptions{Page: 2},
	}
	builds, _, err := client.Builds.List(opt)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_List_by_created_date(t *testing.T) {
	setup()
	defer teardown()

	ts, err := time.Parse(BuildKiteDateFormat, "2016-03-24T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc("/v2/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"created_from": "2016-03-24T01:00:00Z",
			"created_to":   "2016-03-24T02:00:00Z",
		})
		fmt.Fprint(w, `[{"id":"123"}]`)
	})

	opt := &BuildsListOptions{
		CreatedFrom: ts,
		CreatedTo:   ts.Add(time.Hour),
	}
	builds, _, err := client.Builds.List(opt)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_ListByOrg(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	builds, _, err := client.Builds.ListByOrg("my-great-org", nil)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}

func TestBuildsService_ListByPipeline(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/pipelines/sup-keith/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	builds, _, err := client.Builds.ListByPipeline("my-great-org", "sup-keith", nil)
	if err != nil {
		t.Errorf("Builds.List returned error: %v", err)
	}

	want := []Build{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(builds, want) {
		t.Errorf("Builds.List returned %+v, want %+v", builds, want)
	}
}
