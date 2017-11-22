// Copyright 2014 Mark Wolfe. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildkite

import (
	"fmt"
	"io"
)

// ArtifactsService handles communication with the artifact related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api/artifacts
type ArtifactsService struct {
	client *Client
}

// Artifact represents an artifact which has been stored from a build
type Artifact struct {
	ID           *string `json:"id,omitempty"`
	JobID        *string `json:"job_id,omitempty"`
	URL          *string `json:"url,omitempty"`
	DownloadURL  *string `json:"download_url,omitempty"`
	State        *string `json:"state,omitempty"`
	Path         *string `json:"path,omitempty"`
	Dirname      *string `json:"dirname,omitempty"`
	Filename     *string `json:"filename,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
	FileSize     *int64  `json:"file_size,omitempty"`
	GlobPath     *string `json:"glob_path,omitempty"`
	OriginalPath *string `json:"original_path,omitempty"`
	SHA1         *string `json:"sha1sum,omitempty"`
}

// ArtifactListOptions specifies the optional parameters to the
// ArtifactsService.List method.
type ArtifactListOptions struct {
	ListOptions
}

// ListByBuild gets artifacts for a specific build
//
// buildkite API docs: https://buildkite.com/docs/api/artifacts#list-artifacts-for-a-build
func (as *ArtifactsService) ListByBuild(org string, pipeline string, build string, opt *ArtifactListOptions) ([]Artifact, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds/%s/artifacts", org, pipeline, build)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := as.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	artifacts := new([]Artifact)
	resp, err := as.client.Do(req, artifacts)
	if err != nil {
		return nil, resp, err
	}
	return *artifacts, resp, err
}

// ListByBuild gets artifacts for a specific build
//
// buildkite API docs: https://buildkite.com/docs/api/artifacts#list-artifacts-for-a-build
func (as *ArtifactsService) DownloadArtifactByURL(url string, w io.Writer) (*Response, error) {
	req, err := as.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := as.client.Do(req, w)
	if err != nil {
		return resp, err
	}

	return resp, err
}
