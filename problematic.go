package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	. "github.com/gorilla/feeds"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	. "strconv"
	. "strings"
	"time"
)

// Feed formats
const ATOM_FORMAT string = "atom"
const RSS_FORMAT string = "rss"

// Debug mode
var DEBUG bool

// Github personal access token
var token string

func getIssues() (allIssues []github.Issue, err error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
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
	return allIssues, nil
}

func atomViewHandler(responseWriter http.ResponseWriter, request *http.Request) {
	viewHandler(responseWriter, request, ATOM_FORMAT)
}

func rssViewHandler(responseWriter http.ResponseWriter, request *http.Request) {
	viewHandler(responseWriter, request, RSS_FORMAT)
}

func viewHandler(responseWriter http.ResponseWriter, request *http.Request, format string) {
	var now = time.Now()

	if DEBUG {
		log.Println(request.Method, "\""+request.URL.Path+"\"", "-- S:", request.RemoteAddr[:Index(request.RemoteAddr, ":")])
	}

	feed := &Feed{
		Title:       "My Github Issues",
		Link:        &Link{Href: "http://localhost:8888/issues"},
		Description: "My active github issues",
		Author:      &Author{"Problematic.go", "problematic.go"},
		Created:     now,
	}

	var allIssues []github.Issue

	allIssues, err := getIssues()
	if err != nil {
		log.Fatal("Could not retrieve issues from Github. Exiting.")
		os.Exit(1)
	}

	for i := range allIssues {
		issue := allIssues[i]
		user := issue.User
		var issueItem = &Item{
			Id:          *issue.HTMLURL,
			Title:       *issue.Title,
			Link:        &Link{Href: *issue.HTMLURL},
			Description: *issue.Body,
			Author:      &Author{*user.Login, ""},
			Created:     now,
		}
		feed.Items = append(feed.Items, issueItem)
	}

	type formatter func() (string, error)
	var feedFormatter formatter

	if format == RSS_FORMAT {
		feedFormatter = feed.ToRss
	} else if format == ATOM_FORMAT {
		feedFormatter = feed.ToAtom
	}
	feedOutput, err := feedFormatter()
	if err != nil {
		log.Fatal("Failed to generate RSS output for feed data.")
		os.Exit(1)
	}

	fmt.Fprint(responseWriter, feedOutput)
}

func main() {
	// Handle command line arguments
	flag.BoolVar(&DEBUG, "debug", false, "Enable debugging output")
	flag.StringVar(&token, "token", "", "Github Personal Access Token")
	var serverPort = flag.Int("port", 8888, "Port to serve HTTP requsts to.")
	flag.Parse()
	if token == "" {
		log.Fatal("You must provide the token argument. Exiting.")
		os.Exit(1)
	}

	if DEBUG {
		log.Println("Serving at 127.0.0.1:"+Itoa(*serverPort), "|", "Using token: "+token)
	}

	http.HandleFunc("/atom", atomViewHandler)
	http.HandleFunc("/rss", rssViewHandler)
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", *serverPort), nil)
}
