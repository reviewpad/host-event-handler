// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/go-github/v42/github"
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

// processWorkflowRunEvent process GitHub "workflow_run" event.
// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_run
func processWorkflowRunEvent(e *github.WorkflowRunEvent, token string) []int {
	Log("processing 'workflow_run' event")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	prs, err := GetPullRequests(ctx, client, *e.Repo.Owner.Login, *e.Repo.Name, nil)
	if err != nil {
		log.Fatal(err)
	}

	Log("fetched %v prs", len(prs))

	for _, pr := range prs {
		if *pr.Head.SHA == *e.WorkflowRun.HeadSHA {
			Log("found pr %v", *pr.Number)
			return []int{*pr.Number}
		}
	}

	Log("no pr found with the head sha %v", *e.WorkflowRun.HeadSHA)

	return []int{}
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

// processEvent process the GitHub event and returns the list of pull requests that are affected by the event.
func ProcessEvent(event *ActionEvent) []int {
	eventPayload, err := github.ParseWebHook(*event.EventName, *event.EventPayload)
	if err != nil {
		log.Fatal(err)
	}

	switch payload := eventPayload.(type) {
	// Handle github events triggered by actions
	// For more information, visit: https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	case *github.WorkflowRunEvent:
		return processWorkflowRunEvent(payload, *event.Token)
	case *github.PullRequestEvent:
		return processPullRequestEvent(payload, *event.Token)
	case *github.PullRequestTargetEvent:
		return processPullRequestTargetEvent(payload, *event.Token)
	}

	return []int{}
}
