// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/reviewpad/host-event-handler/handler"
)

func getEnvVariable(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		log.Fatal(fmt.Errorf("missing %s env variable", name))
	}
	return value
}

func main() {
	rawEvent := getEnvVariable("INPUT_EVENT")

	event, err := handler.ParseEvent(rawEvent)
	if err != nil {
		log.Fatal(err)
	}

	handler.ProcessEvent(event)
}
