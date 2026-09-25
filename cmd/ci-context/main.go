package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/juli3nk/ci-context/internal/ci"
)

func main() {
	var (
		baseRef      = flag.String("base-ref", "", "Base ref for PRs")
		since        = flag.String("since", "", "Minimum reference commit; base will never be older than this")
		githubOutput = flag.Bool("github-output", false, "Output in GITHUB_OUTPUT format")
		debug        = flag.Bool("debug", false, "Display debug mode")
	)
	flag.Parse()

	if *debug {
		ctx := ci.Detect()

		fmt.Println(ctx.Output())

		return
	}

	refs, err := ci.GetRefs(*baseRef, *since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ci-detect error: %v\n", err)
		os.Exit(1)
	}

	if *githubOutput {
		fmt.Println(refs.GithubOutput())
	} else {
		fmt.Println(refs.Json())
	}
}
