// Copyright 2014 Mark Wolfe. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildkite

import (
	"fmt"
	"time"
)

// BuildsService handles communication with the build related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api/builds
type BuildsService struct {
	client *Client
}

// Author of a commit (used in CreateBuild)
type Author struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// Create a build.
type CreateBuild struct {
	Commit  string `json:"commit"`
	Branch  string `json:"branch"`
	Message string `json:"message"`

	// Optional fields
	Author                      Author            `json:"author,omitempty"`
	Env                         map[string]string `json:"env,omitempty"`
	MetaData                    map[string]string `json:"meta_data,omitempty"`
	IgnorePipelineBranchFilters bool              `json:"ignore_pipeline_branch_filters,omitempty"`
}

// Creator represents who created a build
type Creator struct {
	AvatarURL string     `json:"avatar_url"`
	CreatedAt *Timestamp `json:"created_at"`
	Email     string     `json:"email"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
}

// Build represents a build which has run in buildkite
type Build struct {
	ID          *string                `json:"id,omitempty"`
	URL         *string                `json:"url,omitempty"`
	WebURL      *string                `json:"web_url,omitempty"`
	Number      *int                   `json:"number,omitempty"`
	State       *string                `json:"state,omitempty"`
	Message     *string                `json:"message,omitempty"`
	Commit      *string                `json:"commit,omitempty"`
	Branch      *string                `json:"branch,omitempty"`
	Env         map[string]interface{} `json:"env,omitempty"`
	CreatedAt   *Timestamp             `json:"created_at,omitempty"`
	ScheduledAt *Timestamp             `json:"scheduled_at,omitempty"`
	StartedAt   *Timestamp             `json:"started_at,omitempty"`
	FinishedAt  *Timestamp             `json:"finished_at,omitempty"`
	MetaData    interface{}            `json:"meta_data,omitempty"`
	Creator     *Creator               `json:"creator,omitempty"`

	// jobs run during the build
	Jobs []*Job `json:"jobs,omitempty"`

	// the pipeline this build is associated with
	Pipeline *Pipeline `json:"pipeline,omitempty"`
}

// Job represents a job run during a build in buildkite
type Job struct {
	ID              *string    `json:"id,omitempty"`
	Type            *string    `json:"type,omitempty"`
	Name            *string    `json:"name,omitempty"`
	State           *string    `json:"state,omitempty"`
	LogsURL         *string    `json:"logs_url,omitempty"`
	RawLogsURL      *string    `json:"raw_log_url,omitempty"`
	Command         *string    `json:"command,omitempty"`
	ExitStatus      *int       `json:"exit_status,omitempty"`
	ArtifactPaths   *string    `json:"artifact_paths,omitempty"`
	CreatedAt       *Timestamp `json:"created_at,omitempty"`
	ScheduledAt     *Timestamp `json:"scheduled_at,omitempty"`
	StartedAt       *Timestamp `json:"started_at,omitempty"`
	FinishedAt      *Timestamp `json:"finished_at,omitempty"`
	Agent           Agent      `json:"agent,omitempty"`
	AgentQueryRules []string   `json:"agent_query_rules,omitempty"`
	WebURL          string     `json:"web_url"`
}

// BuildsListOptions specifies the optional parameters to the
// BuildsService.List method.
type BuildsListOptions struct {

	// Filters the results by the user who created the build
	Creator string `url:"creator,omitempty"`

	// Filters the results by builds created on or after the given time
	CreatedFrom time.Time `url:"created_from,omitempty"`

	// Filters the results by builds created before the given time
	CreatedTo time.Time `url:"created_to,omitempty"`

	// Filters the results by builds finished on or after the given time
	FinishedFrom time.Time `url:"finished_from,omitempty"`

	// State of builds to list.  Possible values are: running, scheduled, passed,
	// failed, canceled, skipped and not_run. Default is "".
	State []string `url:"state,brackets,omitempty"`

	// Branch filter by the name of the branch. Default is "".
	Branch string `url:"branch,omitempty"`

	ListOptions
}

func (as *BuildsService) Create(org string, pipeline string, b *CreateBuild) (*Build, *Response, error) {
	u := fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds", org, pipeline)

	req, err := as.client.NewRequest("POST", u, b)
	if err != nil {
		return nil, nil, err
	}

	build := new(Build)
	resp, err := as.client.Do(req, build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, err
}

// Get fetches a build.
//
// buildkite API docs: https://buildkite.com/docs/api/builds#get-a-build
func (as *BuildsService) Get(org string, pipeline string, id string) (*Build, *Response, error) {
	u := fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds/%s", org, pipeline, id)

	req, err := as.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	build := new(Build)
	resp, err := as.client.Do(req, build)
	if err != nil {
		return nil, resp, err
	}

	return build, resp, err
}

// List the builds for the current user.
//
// buildkite API docs: https://buildkite.com/docs/api/builds#list-all-builds
func (bs *BuildsService) List(opt *BuildsListOptions) ([]Build, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/builds")

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	orgs := new([]Build)
	resp, err := bs.client.Do(req, orgs)
	if err != nil {
		return nil, resp, err
	}

	return *orgs, resp, err
}

// ListByOrg lists the builds within the specified orginisation.
//
// buildkite API docs: https://buildkite.com/docs/api/builds#list-builds-for-an-organization
func (bs *BuildsService) ListByOrg(org string, opt *BuildsListOptions) ([]Build, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/builds", org)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	orgs := new([]Build)
	resp, err := bs.client.Do(req, orgs)
	if err != nil {
		return nil, resp, err
	}

	return *orgs, resp, err
}

// ListByPipeline lists the builds for a pipeline within the specified originisation.
//
// buildkite API docs: https://buildkite.com/docs/api/builds#list-builds-for-a-pipeline
func (bs *BuildsService) ListByPipeline(org string, pipeline string, opt *BuildsListOptions) ([]Build, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/pipelines/%s/builds", org, pipeline)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := bs.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	orgs := new([]Build)
	resp, err := bs.client.Do(req, orgs)
	if err != nil {
		return nil, resp, err
	}

	return *orgs, resp, err
}
