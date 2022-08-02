// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/reviewpad/v3/utils"
	"golang.org/x/oauth2"
)

// ParseEvent parses GitHub raw event to ActionEvent
func ParseEvent(rawEvent string) (*ActionEvent, error) {
	event := &ActionEvent{}

	Log("parsing event %v", rawEvent)

	err := json.Unmarshal([]byte(rawEvent), &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func newGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// processWorkflowRunEvent process GitHub "workflow_run" event.
// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_run
func processWorkflowRunEvent(e *github.WorkflowRunEvent, token string) ([]int, error) {
	Log("processing 'workflow_run' event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()
	client := newGithubClient(ctx, token)

	prs, err := utils.GetPullRequests(ctx, client, *e.Repo.Owner.Login, *e.Repo.Name)
	if err != nil {
		return nil, fmt.Errorf("get pull requests: %w", err)
	}

	Log("fetched %v prs", len(prs))

	for _, pr := range prs {
		if *pr.Head.SHA == *e.WorkflowRun.HeadSHA {
			Log("found pr %v", *pr.Number)
			return []int{*pr.Number}, nil
		}
	}

	Log("no pr found with the head sha %v", *e.WorkflowRun.HeadSHA)

	return []int{}, nil
}

// processPullRequestEvent process GitHub "pull_request" event.
// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
func processPullRequestEvent(e *github.PullRequestEvent, token string) []int {
	Log("processing 'pull_request' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []int{*e.PullRequest.Number}
}

// processPullRequestTargetEvent process GitHub "pull_request_target" event.
// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_target
func processPullRequestTargetEvent(e *github.PullRequestTargetEvent, token string) []int {
	Log("processing 'pull_request_target' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []int{*e.PullRequest.Number}
}

// processPullRequestReviewEvent process GitHub "pull_request_review" event.
// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_review
func processPullRequestReviewEvent(e *github.PullRequestReviewEvent, token string) []int {
	Log("processing 'pull_request_review' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []int{*e.PullRequest.Number}
}

func processCronEvent(token string, e *ActionEvent) ([]int, error) {
	Log("processing 'schedule' event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()
	client := newGithubClient(ctx, token)

	repoParts := strings.SplitN(*e.Repository, "/", 2)
	prs, err := utils.GetPullRequests(ctx, client, repoParts[0], repoParts[1])
	if err != nil {
		return nil, fmt.Errorf("get pull requests: %w", err)
	}

	Log("fetched %d prs", len(prs))

	nums := make([]int, 0, len(prs))
	for _, pr := range prs {
		nums = append(nums, *pr.Number)
	}

	Log("found prs %v", nums)

	return nums, nil
}

// processEvent process the GitHub event and returns the list of pull requests that are affected by the event.
func ProcessEvent(event *ActionEvent) ([]int, error) {
	// These events do not have an equivalent in the GitHub webhooks, thus
	// parsing them with github.ParseWebhook would return an error.
	// These are the webhook events: https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
	// And these are the "workflow events": https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	switch *event.EventName {
	case "schedule":
		return processCronEvent(*event.Token, event)
	}

	eventPayload, err := github.ParseWebHook(*event.EventName, *event.EventPayload)
	if err != nil {
		return nil, fmt.Errorf("parse github webhook: %w", err)
	}

	switch payload := eventPayload.(type) {
	// Handle github events triggered by actions
	// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	case *github.WorkflowRunEvent:
		return processWorkflowRunEvent(payload, *event.Token)
	case *github.PullRequestEvent:
		return processPullRequestEvent(payload, *event.Token), nil
	case *github.PullRequestTargetEvent:
		return processPullRequestTargetEvent(payload, *event.Token), nil
	case *github.PullRequestReviewEvent:
		return processPullRequestReviewEvent(payload, *event.Token), nil
	}

	return nil, fmt.Errorf("unknown event payload type: %T", eventPayload)
}
