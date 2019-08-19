package main

import (
	"bufio"
	"context"
	"fmt"
	//"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v27/github"
)

const maxPerPage = 1000

func main() {
	ctx := context.Background()
	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: "... Access token ..."},
	//)
	//tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(nil)

	opt := github.ListOptions {PerPage: maxPerPage, Page: 1}
	var totalIssues []*github.Issue
	for {
		issues, response, err := client.Issues.ListByRepo(ctx, "docker", "compose", &github.IssueListByRepoOptions{
			// Date of the introduction of the {bug, feature request} templates
			Since: time.Date(2018, time.October, 18, 0, 0, 0, 0, time.Local),
			State: "all",
			ListOptions: opt,
		})
		if err != nil {
			log.Printf("%v", err)
			os.Exit(1)
		}
		totalIssues = append(totalIssues, issues...)
		log.Printf("Processing page: %d, totalIssues %d\n", opt.Page, len(totalIssues))
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	green := color.New(color.BgGreen, color.FgWhite).SprintFunc()
	blue := color.New(color.BgBlue, color.FgWhite).SprintFunc()
	red := color.New(color.BgRed, color.FgWhite).SprintFunc()

	bugCount, featureCount, noMatchCount := 0, 0, 0
	for _, issue := range totalIssues {
		var matchColor func(a ...interface{}) string
		var matchText string
		switch matchIssueTemplate(issue.GetBody()) {
		case BUG:
			matchColor = green
			matchText = "  THAT'S A BUG  "
			bugCount++
		case FEATURE:
			matchColor = blue
			matchText = "THAT'S A FEATURE"
			featureCount++
		case NO_MATCH:
			matchColor = red
			matchText = "    NO MATCH    "
			noMatchCount++
		}
		fmt.Printf("#%05d: %s -> %q\n", issue.GetNumber(), matchColor(matchText), issue.GetTitle())
	}
	fmt.Printf("%s: %02.2f%% (%d)\n", green("  BUGS  "), percentage(bugCount, len(totalIssues)), bugCount)
	fmt.Printf("%s: %02.2f%% (%d)\n", blue("FEATURES"), percentage(featureCount, len(totalIssues)), featureCount)
	fmt.Printf("%s: %02.2f%% (%d)\n\n", red("NO MATCH"), percentage(noMatchCount, len(totalIssues)), noMatchCount)
	fmt.Printf("Total of issues: %d\n", len(totalIssues))
}

const(
	BUG = iota
	FEATURE
	NO_MATCH
)

func percentage(x, total int) float32 {
	return float32(x) * 100 / float32(total)
}

func matchIssueTemplate(issueBody string) int32 {
	bugReportMarks := getBugReportMarks()
	featureRequestMarks := getFeatureRequestMarks()
	for _, line := range strings.Split(issueBody,"\n") {
		line = strings.ReplaceAll(line, "\r", "")
		if len(bugReportMarks) > 0 && line == bugReportMarks[0] {
			bugReportMarks = bugReportMarks[1:]
		}
		if len(featureRequestMarks) > 0 && line == featureRequestMarks[0] {
			featureRequestMarks = featureRequestMarks[1:]
		}
		if len(bugReportMarks) == 0 {
			return BUG
		}
		if len(featureRequestMarks) == 0 {
			return FEATURE
		}
	}
	return NO_MATCH
}

func isCheckLine(line string) bool {
	return strings.HasPrefix(line, "##") || strings.HasPrefix(line, "**")
}

func getBugReportMarks() []string {
	return getMarks(".github/ISSUE_TEMPLATE/bug_report.md")
}

func getFeatureRequestMarks() []string {
	return getMarks(".github/ISSUE_TEMPLATE/feature_request.md")
}

func getMarks(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var linesToCheck []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if isCheckLine(line) {
			linesToCheck = append(linesToCheck, strings.Split(line, "\n")[0])
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return linesToCheck
}
