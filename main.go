package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const working = "Currently working on:"
const done = "Done since last week:"

var repos = []string{"redhat-developer/odo"}

func getPRs(client *github.Client, ctx context.Context, repo string, updated string, is string) ([]github.Issue, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allIssues []github.Issue
	for {
		query := fmt.Sprintf("repo:%s is:pr updated:>=%s is:%s", repo, updated, is)
		issues, resp, err := client.Search.Issues(ctx, query, opt)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allIssues, nil

}

func main() {
	token := os.Getenv("GITHUBREPORT_TOKEN")

	ctx := context.Background()

	var client *github.Client
	// if token is set use it
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)

		client = github.NewClient(tc)
	} else {
		// otherwise use anonymous client
		client = github.NewClient(nil)
	}

	// one week ago
	date := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	allInfo := make(map[string]map[string][]string)
	allInfo[working] = make(map[string][]string)
	allInfo[done] = make(map[string][]string)

	for _, repo := range repos {
		openPRs, err := getPRs(client, ctx, repo, date, "open")
		if err != nil {
			panic(err)
		}
		closePRs, err := getPRs(client, ctx, repo, date, "closed")
		if err != nil {
			panic(err)
		}

		for _, pr := range openPRs {
			allInfo[working][repo] = append(allInfo[working][repo], fmt.Sprintf("[%s](%s)", *pr.Title, *pr.HTMLURL))
		}
		for _, pr := range closePRs {
			allInfo[done][repo] = append(allInfo[done][repo], fmt.Sprintf("[%s](%s)", *pr.Title, *pr.HTMLURL))
		}

	}

	for info, repos := range allInfo {
		fmt.Printf("- %s \n", info)
		for repo, prs := range repos {
			fmt.Printf("  - %s\n", strings.Split(repo, "/")[1])
			for _, pr := range prs {
				fmt.Printf("    - %s\n", pr)
			}
		}
	}
}
