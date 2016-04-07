// Copyright 2014 Mark Wolfe. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildkite

import "fmt"

// Emoji emoji, what else can you say?
type Emoji struct {
	Name *string `json:"name,omitempty"`
	URL  *string `json:"url,omitempty"`
}

// ListEmojis list all the emojis for a given account, including custom emojis and aliases.
//
// buildkite API docs: https://buildkite.com/docs/api/emojis
func (c *Client) ListEmojis(org string) ([]Emoji, *Response, error) {

	var u string

	u = fmt.Sprintf("v2/organizations/%s/emojis", org)

	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	emoji := new([]Emoji)
	resp, err := c.Do(req, emoji)
	if err != nil {
		return nil, resp, err
	}

	return *emoji, resp, nil
}

// Token an oauth access token for the buildkite service
type Token struct {
	AccessToken *string `json:"access_token,omitempty"`
	Type        *string `json:"token_type,omitempty"`
}
