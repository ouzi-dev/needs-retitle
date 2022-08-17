/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const (
	// PluginName is the name of this plugin
	PluginName                 = "needs-retitle"
	defaultNeedsRetitleMessage = "Wrong title for PR, allowed titles need to match the regular expression: %s"
	needsRetitleLabel          = "needs-retitle"
)

var sleep = time.Sleep

type githubClient interface {
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	CreateComment(org, repo string, number int, comment string) error
	BotUser() (*github.UserData, error)
	AddLabel(org, repo string, number int, label string) error
	RemoveLabel(org, repo string, number int, label string) error
	DeleteStaleComments(org, repo string, number int, comments []github.IssueComment, isStale func(github.IssueComment) bool) error
	QueryWithGitHubAppsSupport(ctx context.Context, q interface{}, vars map[string]interface{}, org string) error
	GetPullRequest(org, repo string, number int) (*github.PullRequest, error)
}

type Plugin struct {
	mut sync.Mutex
	c   *pluginConfig
}

type pluginConfig struct {
	errorMessage string
	re           *regexp.Regexp
}

// HelpProvider constructs the PluginHelp for this plugin that takes into account enabled repositories.
// HelpProvider defines the type for function that construct the PluginHelp for plugins.
func HelpProvider(_ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
			Description: `The ` + PluginName + ` plugin manages the '` + needsRetitleLabel + `' label by removing it from Pull Requests with a title that matches the configured regular expression and adding it to those which doesn't.
The plugin reacts to commit changes on PRs in addition to periodically scanning all open PRs for any changes in the titles.`,
		},
		nil
}

func (p *Plugin) SetConfig(m string, r *regexp.Regexp) {
	p.mut.Lock()
	defer p.mut.Unlock()

	errorMessage := fmt.Sprintf(defaultNeedsRetitleMessage, r.String())
	if len(m) > 0 {
		errorMessage = m
	}

	p.c = &pluginConfig{
		errorMessage: errorMessage,
		re:           r,
	}
}

func (p *Plugin) GetConfig() *pluginConfig {
	p.mut.Lock()
	defer p.mut.Unlock()
	return p.c
}

// HandlePullRequestEvent handles a GitHub pull request event and adds or removes a
// "needs-retitle" label based on whether the title matches the provided regular expression
func (p *Plugin) HandlePullRequestEvent(log *logrus.Entry, ghc githubClient, pre *github.PullRequestEvent) error {
	if pre.Action != github.PullRequestActionOpened &&
		pre.Action != github.PullRequestActionSynchronize &&
		pre.Action != github.PullRequestActionReopened &&
		pre.Action != github.PullRequestActionEdited {
		return nil
	}

	return p.handle(log, ghc, &pre.PullRequest)
}

// HandleIssueCommentEvent handles a GitHub issue comment event and adds or removes a
// "needs-retitle" label if the title matches the provided regular expression or not
func (p *Plugin) HandleIssueCommentEvent(log *logrus.Entry, ghc githubClient, ice *github.IssueCommentEvent) error {
	if !ice.Issue.IsPullRequest() {
		return nil
	}
	pr, err := ghc.GetPullRequest(ice.Repo.Owner.Login, ice.Repo.Name, ice.Issue.Number)
	if err != nil {
		return err
	}

	return p.handle(log, ghc, pr)
}

// handle handles a GitHub PR to determine if the "needs-retitle"
// label needs to be added or removed.
func (p *Plugin) handle(log *logrus.Entry, ghc githubClient, pr *github.PullRequest) error {
	if pr.Merged {
		return nil
	}

	c := p.GetConfig()

	if c == nil {
		log.Warnf("No regular expression provided, please check your settings")
		return nil
	}

	org := pr.Base.Repo.Owner.Login
	repo := pr.Base.Repo.Name
	number := pr.Number
	title := pr.Title

	issueLabels, err := ghc.GetIssueLabels(org, repo, number)
	if err != nil {
		return err
	}
	hasLabel := github.HasLabel(needsRetitleLabel, issueLabels)

	return p.takeAction(log, ghc, org, repo, number, pr.User.Login, hasLabel, title, c)
}

// HandleAll checks all orgs and repos that enabled this plugin for open PRs to
// determine if the "needs-retitle" label needs to be added or removed.
func (p *Plugin) HandleAll(log *logrus.Entry, ghc githubClient, config *plugins.Configuration) error {
	log.Info("Checking all PRs.")

	c := p.GetConfig()

	if c == nil {
		log.Warnf("No regular expression provided, please check your settings")
		return nil
	}

	orgs, repos := config.EnabledReposForExternalPlugin(PluginName)
	if len(orgs) == 0 && len(repos) == 0 {
		log.Warnf("No repos have been configured for the %s plugin", PluginName)
		return nil
	}
	var buf bytes.Buffer
	fmt.Fprint(&buf, "archived:false is:pr is:open")
	for _, org := range orgs {
		fmt.Fprintf(&buf, " org:\"%s\"", org)
	}
	for _, repo := range repos {
		fmt.Fprintf(&buf, " repo:\"%s\"", repo)
	}
	prs, err := search(context.Background(), log, ghc, buf.String())
	if err != nil {
		return err
	}
	log.Infof("Considering %d PRs.", len(prs))

	for _, pr := range prs {
		org := string(pr.Repository.Owner.Login)
		repo := string(pr.Repository.Name)
		num := int(pr.Number)
		title := string(pr.Title)
		l := log.WithFields(logrus.Fields{
			"org":  org,
			"repo": repo,
			"pr":   num,
		})
		hasLabel := false
		for _, label := range pr.Labels.Nodes {
			if label.Name == needsRetitleLabel {
				hasLabel = true
				break
			}
		}
		err := p.takeAction(
			l,
			ghc,
			org,
			repo,
			num,
			string(pr.Author.Login),
			hasLabel,
			title,
			c,
		)
		if err != nil {
			l.WithError(err).Error("Error handling PR.")
		}
	}
	return nil
}

// takeAction adds or removes the "needs-rebase" label based on the current
// state of the PR (hasLabel and mergeable). It also handles adding and
// removing GitHub comments notifying the PR author that a rebase is needed.
func (p *Plugin) takeAction(log *logrus.Entry, ghc githubClient, org, repo string, num int, author string, hasLabel bool, title string, c *pluginConfig) error {
	needsRetitleMessage := c.errorMessage
	titleOk := c.re.MatchString(title)
	if !titleOk && !hasLabel {
		if err := ghc.AddLabel(org, repo, num, needsRetitleLabel); err != nil {
			log.WithError(err).Errorf("Failed to add %q label.", needsRetitleLabel)
		}
		msg := plugins.FormatSimpleResponse(author, needsRetitleMessage)
		return ghc.CreateComment(org, repo, num, msg)
	} else if titleOk && hasLabel {
		// remove label and prune comment
		if err := ghc.RemoveLabel(org, repo, num, needsRetitleLabel); err != nil {
			log.WithError(err).Errorf("Failed to remove %q label.", needsRetitleLabel)
		}
		botUser, err := ghc.BotUser()
		botName := botUser.Name
		if err != nil {
			return err
		}
		return ghc.DeleteStaleComments(org, repo, num, nil, shouldPrune(botName, needsRetitleMessage))
	}
	return nil
}

func shouldPrune(botName string, msg string) func(github.IssueComment) bool {
	return func(ic github.IssueComment) bool {
		return github.NormLogin(botName) == github.NormLogin(ic.User.Login) &&
			strings.Contains(ic.Body, msg)
	}
}

func search(ctx context.Context, log *logrus.Entry, ghc githubClient, q string) ([]pullRequest, error) {
	var ret []pullRequest
	vars := map[string]interface{}{
		"query":        githubql.String(q),
		"searchCursor": (*githubql.String)(nil),
	}
	var totalCost int
	var remaining int
	for {
		sq := searchQuery{}
		if err := ghc.QueryWithGitHubAppsSupport(ctx, &sq, vars, ""); err != nil {
			return nil, err
		}
		totalCost += int(sq.RateLimit.Cost)
		remaining = int(sq.RateLimit.Remaining)
		for _, n := range sq.Search.Nodes {
			ret = append(ret, n.PullRequest)
		}
		if !sq.Search.PageInfo.HasNextPage {
			break
		}
		vars["searchCursor"] = githubql.NewString(sq.Search.PageInfo.EndCursor)
	}
	log.Infof("Search for query \"%s\" cost %d point(s). %d remaining.", q, totalCost, remaining)
	return ret, nil
}

type pullRequest struct {
	Number githubql.Int
	Title  githubql.String
	Author struct {
		Login githubql.String
	}
	Repository struct {
		Name  githubql.String
		Owner struct {
			Login githubql.String
		}
	}
	Labels struct {
		Nodes []struct {
			Name githubql.String
		}
	} `graphql:"labels(first:100)"`
}

type searchQuery struct {
	RateLimit struct {
		Cost      githubql.Int
		Remaining githubql.Int
	}
	Search struct {
		PageInfo struct {
			HasNextPage githubql.Boolean
			EndCursor   githubql.String
		}
		Nodes []struct {
			PullRequest pullRequest `graphql:"... on PullRequest"`
		}
	} `graphql:"search(type: ISSUE, first: 100, after: $searchCursor, query: $query)"`
}
