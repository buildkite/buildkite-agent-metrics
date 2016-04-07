package buildkite

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the buildkite client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a buildkite.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// github client configured to use test server
	client = NewClient(http.DefaultClient)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string
type valuesList []struct{ key, val string }

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func testFormValuesList(t *testing.T, r *http.Request, values valuesList) {
	want := url.Values{}
	for _, v := range values {
		want.Add(v.key, v.val)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)

	if got, want := c.BaseURL.String(), defaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	c := NewClient(nil)
	inURL, outURL := "/foo", defaultBaseURL+"foo"
	inBody := &User{ID: String("123"), Name: String("Jane Doe"), Email: String("jane@doe.com")}
	outBody := `{"id":"123","name":"Jane Doe","email":"jane@doe.com"}` + "\n"

	req, _ := c.NewRequest("GET", inURL, inBody)

	// test that relative URL was expanded
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	// test that body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%q) Body is %v, want %v", inBody, got, want)
	}

	// test that default user-agent is attached to the request
	if got, want := req.Header.Get("User-Agent"), c.UserAgent; got != want {
		t.Errorf("NewRequest() User-Agent is %v, want %v", got, want)
	}
}

func TestResponse_populatePageValues(t *testing.T) {
	r := http.Response{
		Header: http.Header{
			"Link": {`<https://api.buildkite.com/?page=1>; rel="first",` +
				` <https://api.buildkite.com/?page=2>; rel="prev",` +
				` <https://api.buildkite.com/?page=4>; rel="next",` +
				` <https://api.buildkite.com/?page=5>; rel="last"`,
			},
		},
	}

	response := newResponse(&r)
	if got, want := response.FirstPage, 1; got != want {
		t.Errorf("response.FirstPage: %v, want %v", got, want)
	}
	if got, want := response.PrevPage, 2; want != got {
		t.Errorf("response.PrevPage: %v, want %v", got, want)
	}
	if got, want := response.NextPage, 4; want != got {
		t.Errorf("response.NextPage: %v, want %v", got, want)
	}
	if got, want := response.LastPage, 5; want != got {
		t.Errorf("response.LastPage: %v, want %v", got, want)
	}
}
