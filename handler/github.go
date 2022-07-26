// Copyright (C) 2022 Explore.dev, Unipessoal Lda - All Rights Reserved
// Use of this source code is governed by a license that can be
// Proprietary and confidential

package handler

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/tomnomnom/linkheader"
)

const maxPerPage int = 100

func PaginatedRequest(
	initFn func() interface{},
	reqFn func(interface{}, int) (interface{}, *github.Response, error),
) (interface{}, error) {
	page := 1
	results, resp, err := reqFn(initFn(), page)
	if err != nil {
		return nil, err
	}

	numPages := ParseNumPages(resp)
	page++
	for page <= numPages && resp.NextPage > page {
		results, _, err = reqFn(results, page)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func ParseNumPagesFromLink(link string) int {
	urlInfo := linkheader.Parse(link).FilterByRel("last")
	if len(urlInfo) < 1 {
		return 0
	}

	urlData, err := url.Parse(urlInfo[0].URL)
	if err != nil {
		return 0
	}

	numPagesStr := urlData.Query().Get("page")
	if numPagesStr == "" {
		return 0
	}

	numPages, err := strconv.ParseInt(numPagesStr, 10, 32)
	if err != nil {
		return 0
	}

	return int(numPages)
}

//ParseNumPages Given a link header string representing pagination info, returns total number of pages.
func ParseNumPages(resp *github.Response) int {
	link := resp.Header.Get("Link")
	if strings.Trim(link, " ") == "" {
		return 0
	}

	return ParseNumPagesFromLink(link)
}

// GetPullRequests fetches all of the open pull requests from the specified repository.
func GetPullRequests(ctx context.Context, client *github.Client, owner string, repo string) ([]*github.PullRequest, error) {
	prs, err := PaginatedRequest(
		func() interface{} {
			return []*github.PullRequest{}
		},
		func(i interface{}, page int) (interface{}, *github.Response, error) {
			allPrs := i.([]*github.PullRequest)
			prs, resp, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: maxPerPage,
				},
			})
			if err != nil {
				return nil, nil, err
			}
			allPrs = append(allPrs, prs...)
			return allPrs, resp, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return prs.([]*github.PullRequest), nil
}
