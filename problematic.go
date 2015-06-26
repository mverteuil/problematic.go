package main

import (
	"fmt"
	"github.com/google/go-github/github"
	. "github.com/gorilla/feeds"
	"golang.org/x/oauth2"
	"os"
	"time"
)

func getIssues() (allIssues []github.Issue, err error) {
	if pat := os.Getenv("PROBLEMATIC_GHPAT"); pat != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: pat})
		tokenClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
		githubClient := github.NewClient(tokenClient)

		opt := &github.IssueListOptions{
			ListOptions: github.ListOptions{PerPage: 10, Page: 0},
		}
		for {
			issues, resp, err := githubClient.Issues.List(true, nil)
			if err != nil {
				return nil, err
			}
			allIssues = append(allIssues, issues...)
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
	}
	return allIssues, nil
}

func main() {
	var now = time.Now()

	feed := &Feed{
		Title:       "My Github Issues",
		Link:        &Link{Href: "http://localhost/issues"},
		Description: "My active github issues",
		Author:      &Author{"Problematic.go", "problematic.go"},
		Created:     now,
		Copyright:   "Your Mom",
	}

	var allIssues []github.Issue

	allIssues, err := getIssues()
	if err != nil {
		panic("I'VE FAILED YOU")
	}

	for i := range allIssues {
		issue := allIssues[i]
		user := issue.User
		var issueItem = &Item{
			Title:       *issue.Title,
			Link:        &Link{Href: *issue.URL},
			Description: *issue.Body,
			Author:      &Author{*user.Login, ""},
			Created:     now,
		}
		feed.Items = append(feed.Items, issueItem)
	}

	atom, err := feed.ToAtom()
	fmt.Printf("%v\r\n", atom)
}
