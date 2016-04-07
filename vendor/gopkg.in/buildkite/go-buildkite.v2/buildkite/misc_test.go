package buildkite

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListEmojis(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations/my-great-org/emojis", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"name":"rocket","url":"https://a.buildboxassets.com/assets/emoji2/unicode/1f680.png?v2"}]`)
	})

	emoji, _, err := client.ListEmojis("my-great-org")
	if err != nil {
		t.Errorf("ListEmojis returned error: %v", err)
	}

	want := []Emoji{{Name: String("rocket"), URL: String("https://a.buildboxassets.com/assets/emoji2/unicode/1f680.png?v2")}}
	if !reflect.DeepEqual(want, emoji) {
		t.Errorf("ListEmojis returned %+v, want %+v", emoji, want)
	}
}
