// Copyright 2014 Mark Wolfe. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildkite

import "fmt"

// PipelinesService handles communication with the pipeline related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api/pipelines
type PipelinesService struct {
	client *Client
}

// Pipeline represents a buildkite pipeline.
type Pipeline struct {
	ID         *string    `json:"id,omitempty"`
	URL        *string    `json:"url,omitempty"`
	WebURL     *string    `json:"web_url,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Slug       *string    `json:"slug,omitempty"`
	Repository *string    `json:"repository,omitempty"`
	BuildsURL  *string    `json:"builds_url,omitempty"`
	BadgeURL   *string    `json:"badge_url,omitempty"`
	CreatedAt  *Timestamp `json:"created_at,omitempty"`

	ScheduledBuildsCount *int `json:"scheduled_builds_count,omitempty"`
	RunningBuildsCount   *int `json:"running_builds_count,omitempty"`
	ScheduledJobsCount   *int `json:"scheduled_jobs_count,omitempty"`
	RunningJobsCount     *int `json:"running_jobs_count,omitempty"`
	WaitingJobsCount     *int `json:"waiting_jobs_count,omitempty"`

	// the provider of sources
	Provider *Provider `json:"provider,omitempty"`

	// build steps
	Steps []*Step `json:"steps,omitempty"`
}

// Provider represents a source code provider.
type Provider struct {
	ID         *string `json:"id,omitempty"`
	WebhookURL *string `json:"webhook_url,omitempty"`
}

// Step represents a build step in buildkites build pipeline
type Step struct {
	Type                *string           `json:"type,omitempty"`
	Name                *string           `json:"name,omitempty"`
	Command             *string           `json:"command,omitempty"`
	ArtifactPaths       *string           `json:"artifact_paths,omitempty"`
	BranchConfiguration *string           `json:"branch_configuration,omitempty"`
	Env                 map[string]string `json:"env,omitempty"`
	TimeoutInMinutes    interface{}       `json:"timeout_in_minutes,omitempty"` // *shrug*
	AgentQueryRules     interface{}       `json:"agent_query_rules,omitempty"`  // *shrug*
}

// PipelineListOptions specifies the optional parameters to the
// PipelinesService.List method.
type PipelineListOptions struct {
	ListOptions
}

// List the pipelines for a given orginisation.
//
// buildkite API docs: https://buildkite.com/docs/api/pipelines#list-pipelines
func (ps *PipelinesService) List(org string, opt *PipelineListOptions) ([]Pipeline, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/pipelines", org)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := ps.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	pipelines := new([]Pipeline)
	resp, err := ps.client.Do(req, pipelines)
	if err != nil {
		return nil, resp, err
	}

	return *pipelines, resp, err
}
