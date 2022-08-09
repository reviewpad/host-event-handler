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

	"github.com/google/go-github/v45/github"
	"github.com/reviewpad/reviewpad/v3/utils"
	"golang.org/x/oauth2"
)

const (
	PullRequest EventKind = "pull_request"
	Issue       EventKind = "issue"
)

type EventKind string

type EventInfo struct {
	Kind   EventKind
	Number int
}

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

func processCronEvent(token string, e *ActionEvent) ([]*EventInfo, error) {
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

	events := make([]*EventInfo, 0)
	for _, pr := range prs {
		events = append(events, &EventInfo{
			Kind:   PullRequest,
			Number: *pr.Number,
		})
	}

	Log("found events %v", events)

	return events, nil
}

func processIssuesEvent(e *github.IssuesEvent) []*EventInfo {
	Log("processing 'issues' event")
	Log("found issue %v", *e.Issue.Number)

	return []*EventInfo{
		{
			Kind:   Issue,
			Number: *e.Issue.Number,
		},
	}
}

func processIssueCommentEvent(e *github.IssueCommentEvent) []*EventInfo {
	Log("processing 'issue_comment' event")
	Log("found issue %v", *e.Issue.Number)

	return []*EventInfo{
		{
			Kind:   Issue,
			Number: *e.Issue.Number,
		},
	}
}

func processPullRequestEvent(e *github.PullRequestEvent) []*EventInfo {
	Log("processing 'pull_request' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []*EventInfo{
		{
			Kind:   PullRequest,
			Number: *e.PullRequest.Number,
		},
	}
}

func processPullRequestReviewEvent(e *github.PullRequestReviewEvent) []*EventInfo {
	Log("processing 'pull_request_review' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []*EventInfo{
		{
			Kind:   PullRequest,
			Number: *e.PullRequest.Number,
		},
	}
}

func processPullRequestReviewCommentEvent(e *github.PullRequestReviewCommentEvent) []*EventInfo {
	Log("processing 'pull_request_review_comment' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []*EventInfo{
		{
			Kind:   PullRequest,
			Number: *e.PullRequest.Number,
		},
	}
}

func processPullRequestTargetEvent(e *github.PullRequestTargetEvent) []*EventInfo {
	Log("processing 'pull_request_target' event")
	Log("found pr %v", *e.PullRequest.Number)

	return []*EventInfo{
		{
			Kind:   PullRequest,
			Number: *e.PullRequest.Number,
		},
	}
}

func processStatusEvent(token string, e *github.StatusEvent) ([]*EventInfo, error) {
	Log("processing 'status' event")

	ctx, canc := context.WithTimeout(context.Background(), time.Minute*10)
	defer canc()
	client := newGithubClient(ctx, token)

	prs, err := utils.GetPullRequests(ctx, client, *e.Repo.Owner.Login, *e.Repo.Name)
	if err != nil {
		return nil, fmt.Errorf("get pull requests: %w", err)
	}

	Log("fetched %v prs", len(prs))

	for _, pr := range prs {
		if *pr.Head.SHA == *e.SHA {
			Log("found pr %v", *pr.Number)
			return []*EventInfo{
				{
					Kind:   PullRequest,
					Number: *pr.Number,
				},
			}, nil
		}
	}

	Log("no pr found with the head sha %v", *e.SHA)

	return []*EventInfo{}, nil
}

func processWorkflowRunEvent(token string, e *github.WorkflowRunEvent) ([]*EventInfo, error) {
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
			return []*EventInfo{
				{
					Kind:   PullRequest,
					Number: *pr.Number,
				},
			}, nil
		}
	}

	Log("no pr found with the head sha %v", *e.WorkflowRun.HeadSHA)

	return []*EventInfo{}, nil
}

// reviewpad-an: critical
// output: the list of pull requests/issues that are affected by the event.
func ProcessEvent(event *ActionEvent) ([]*EventInfo, error) {
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
	case *github.IssuesEvent:
		return processIssuesEvent(payload), nil
	case *github.IssueCommentEvent:
		return processIssueCommentEvent(payload), nil
	case *github.PullRequestEvent:
		return processPullRequestEvent(payload), nil
	case *github.PullRequestReviewEvent:
		return processPullRequestReviewEvent(payload), nil
	case *github.PullRequestReviewCommentEvent:
		return processPullRequestReviewCommentEvent(payload), nil
	case *github.PullRequestTargetEvent:
		return processPullRequestTargetEvent(payload), nil
	case *github.StatusEvent:
		return processStatusEvent(*event.Token, payload)
	case *github.WorkflowRunEvent:
		return processWorkflowRunEvent(*event.Token, payload)
	}

	return nil, fmt.Errorf("unknown event payload type: %T", eventPayload)
}
