// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/radius/cmd/cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run cmd/docgen/main.go <output directory>")
	}

	output := os.Args[1]
	_, err := os.Stat(output)
	if os.IsNotExist(err) {
		err = os.Mkdir(output, 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(cmd.RootCmd, output, frontmatter, link)
	if err != nil {
		log.Fatal(err)
	}
}

const template = `---
type: docs
date: %s
title: "%s CLI reference"
linkTitle: "%s"
slug: %s
url: %s
description: "Details on the %s Radius CLI command"
---
`

func frontmatter(filename string) string {
	now := time.Now().Format(time.RFC3339)
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	command := strings.Replace(base, "_", " ", -1)
	url := "/reference/cli/" + strings.ToLower(base) + "/"
	return fmt.Sprintf(template, now, command, command, base, url, command)
}

func link(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "{{< ref " + strings.ToLower(base) + ".md >}}"
}

type options struct {
}

func visit(cmd *cobra.Command, opt options) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := visit(c, opt); err != nil {
			return err
		}
	}

	return nil
}
