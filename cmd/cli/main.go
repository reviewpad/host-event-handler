// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/reviewpad/host-event-handler/handler"
)

var (
	gitHubToken   = flag.String("github-token", "", "GitHub Personal Access Token (PAT)")
	eventFilePath = flag.String("event-payload", "", "File path to github action event")
)

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "help" {
		usage()
	}

	if *gitHubToken == "" {
		log.Printf("missing argument token")
		usage()
	}

	if *eventFilePath == "" {
		log.Printf("missing argument event")
		usage()
	}

	content, err := ioutil.ReadFile(*eventFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Replace GitHub token on event by PAT
	rawEvent := strings.Replace(string(content), "***", *gitHubToken, 1)

	event, err := handler.ParseEvent(rawEvent)
	if err != nil {
		log.Fatal(err)
	}

	handler.ProcessEvent(event)
}
