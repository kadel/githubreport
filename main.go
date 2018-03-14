package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const working = "Currently working on:"
const done = "Done since last week:"

var repos = []string{"kedgeproject/kedge", "kubernetes/kompose", "redhat-developer/ocdev"}

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

func handler(w http.ResponseWriter, r *http.Request) {
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

	out := fmt.Sprintf("## Report  from:%s to:%s\n", date, time.Now().Format("2006-01-02"))
	for info, repos := range allInfo {
		out = out + fmt.Sprintf("\n\n- %s\n", info)
		for repo, prs := range repos {
			out = out + fmt.Sprintf("    - %s\n", strings.Split(repo, "/")[1])
			for _, pr := range prs {
				out = out + fmt.Sprintf("        - %s\n", pr)
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, string(blackfriday.Run([]byte(out))[:]))

}

func main() {

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
