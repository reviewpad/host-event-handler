// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v42/github"
	"github.com/jarcoal/httpmock"
	"github.com/reviewpad/host-event-handler/handler"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
	"github.com/stretchr/testify/assert"
)

func TestProcessEvent(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	owner := "reviewpad"
	repo := "reviewpad"
	httpmock.RegisterResponder("GET", fmt.Sprintf("https://api.github.com/repos/%v/%v/pulls", owner, repo),
		func(req *http.Request) (*http.Response, error) {
			b, err := json.Marshal([]*github.PullRequest{
				{
					Number: github.Int(aladino.DefaultMockPrNum),
					Head: &github.PullRequestBranch{
						SHA: github.String("4bf24cc72f3a62423927a0ac8d70febad7c78e0g"),
					},
				},
				{
					Number: github.Int(130),
					Head: &github.PullRequestBranch{
						SHA: github.String("4bf24cc72f3a62423927a0ac8d70febad7c78e0k"),
					},
				},
			})
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, b)

			return resp, nil
		},
	)

	buildPayload := func(payload []byte) *json.RawMessage {
		rawPayload := json.RawMessage(payload)
		return &rawPayload
	}

	tests := map[string]struct {
		event   *handler.ActionEvent
		wantVal []int
	}{
		"pull_request": {
			event: &handler.ActionEvent{
				EventName: github.String("pull_request"),
				Token:     github.String("test-token"),
				EventPayload: buildPayload([]byte(`{
					"action": "opened",
					"number": 130,
					"pull_request": {
						"body": "## Description",
						"number": 130
					}
				}`)),
			},
			wantVal: []int{130},
		},
		"pull_request_target": {
			event: &handler.ActionEvent{
				EventName: github.String("pull_request_target"),
				Token:     github.String("test-token"),
				EventPayload: buildPayload([]byte(`{
					"action": "opened",
					"number": 130,
					"pull_request": {
						"body": "## Description",
						"number": 130
					}
				}`)),
			},
			wantVal: []int{130},
		},
		"workflow_run_match": {
			event: &handler.ActionEvent{
				EventName: github.String("workflow_run"),
				Token:     github.String("test-token"),
				EventPayload: buildPayload([]byte(`{
					"action": "completed",
					"repository": {
						"name": "reviewpad",
						"owner": {
							"login": "reviewpad"
						}
					},
					"workflow_run": {
						"head_sha": "4bf24cc72f3a62423927a0ac8d70febad7c78e0g"
					}
				}`)),
			},
			wantVal: []int{aladino.DefaultMockPrNum},
		},
		"workflow_run_no_match": {
			event: &handler.ActionEvent{
				EventName: github.String("workflow_run"),
				Token:     github.String("test-token"),
				EventPayload: buildPayload([]byte(`{
					"action": "completed",
					"repository": {
						"name": "reviewpad",
						"owner": {
							"login": "reviewpad"
						}
					},
					"workflow_run": {
						"head_sha": "4bf24cc72f3a62423927a0ac8d70febad7c78e0a"
					}
				}`)),
			},
			wantVal: []int{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotVal, err := handler.ProcessEvent(test.event)

			assert.Nil(t, err)
			assert.ElementsMatch(t, test.wantVal, gotVal)
		})
	}
}
