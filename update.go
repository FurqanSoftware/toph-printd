package main

import (
	"context"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v53/github"
	"golang.org/x/mod/semver"
)

func checkUpdate(ctx context.Context) error {
	if version == "" || version == "devel" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	version := "v" + version

	if !semver.IsValid(version) {
		return nil
	}

	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	if err != nil {
		return err
	}
	if rel.TagName == nil {
		return nil
	}

	if semver.Compare(*rel.TagName, version) > 0 {
		log.Printf("%s Update available (%s)", color.HiYellowString("[W]"), *rel.TagName)
	}

	return nil
}
