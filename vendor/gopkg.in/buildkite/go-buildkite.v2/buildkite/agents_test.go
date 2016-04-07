package buildkite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestAgentsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/agents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":"123"},{"id":"1234"}]`)
	})

	agents, _, err := client.Agents.List("my-great-org", nil)
	if err != nil {
		t.Errorf("Agents.List returned error: %v", err)
	}

	want := []Agent{{ID: String("123")}, {ID: String("1234")}}
	if !reflect.DeepEqual(agents, want) {
		t.Errorf("Agents.List returned %+v, want %+v", agents, want)
	}
}

func TestAgentsService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/agents/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":"123"}`)
	})

	agent, _, err := client.Agents.Get("my-great-org", "123")
	if err != nil {
		t.Errorf("Agents.Get returned error: %v", err)
	}

	want := &Agent{ID: String("123")}
	if !reflect.DeepEqual(agent, want) {
		t.Errorf("Agents.Get returned %+v, want %+v", agent, want)
	}
}

func TestAgentsService_Create(t *testing.T) {
	setup()
	defer teardown()

	input := &Agent{Name: String("new_agent_bob")}

	mux.HandleFunc("/v2/organizations/my-great-org/agents", func(w http.ResponseWriter, r *http.Request) {
		v := new(Agent)
		json.NewDecoder(r.Body).Decode(&v)

		testMethod(t, r, "POST")

		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":"123"}`)
	})

	agent, _, err := client.Agents.Create("my-great-org", input)
	if err != nil {
		t.Errorf("Agents.Create returned error: %v", err)
	}

	want := &Agent{ID: String("123")}
	if !reflect.DeepEqual(agent, want) {
		t.Errorf("Agents.Create returned %+v, want %+v", agent, want)
	}

}

func TestAgentsService_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/agents/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Agents.Delete("my-great-org", "123")
	if err != nil {
		t.Errorf("Agents.Delete returned error: %v", err)
	}
}
