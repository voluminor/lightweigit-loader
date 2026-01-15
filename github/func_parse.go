package github

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// // // // // // // // // // // // // // // //

var githubNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func parseScpLikeGitHub(s string) (string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected SSH format: %q", s)
	}

	left := parts[0]
	if !strings.HasSuffix(strings.ToLower(left), "@github.com") {
		return "", fmt.Errorf("not GitHub SSH addresses: %q", s)
	}

	path := parts[1]
	owner, repo, err := ownerRepoFromPath(path)
	if err != nil {
		return "", err
	}

	if !githubNameRe.MatchString(owner) || !githubNameRe.MatchString(repo) {
		return "", fmt.Errorf("incorrect owner/repo: %q/%q", owner, repo)
	}

	return owner + "/" + repo, nil
}

func ownerRepoFromPath(p string) (string, string, error) {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return "", "", errors.New("there is no repository path in the url")
	}

	segments := strings.Split(trimmed, "/")
	if len(segments) < 2 {
		return "", "", fmt.Errorf("expected a path of the type /owner/repo, received: %q", p)
	}

	owner := segments[0]
	repo := segments[1]
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner or repo is empty in the path: %q", p)
	}

	return owner, repo, nil
}

// // // //

func Parse(raw string) (*Obj, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, errors.New("an empty URL string")
	}

	if strings.Contains(s, "@github.com:") && !strings.Contains(s, "://") {
		name, err := parseScpLikeGitHub(s)
		return &Obj{name: name}, err
	}

	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("URL could not be parsed: %w", err)
	}

	host := strings.ToLower(u.Hostname())
	if host == "www.github.com" {
		host = "github.com"
	}
	if host != "github.com" {
		return nil, fmt.Errorf("not GitHub host: %q", host)
	}

	owner, repo, err := ownerRepoFromPath(u.Path)
	if err != nil {
		return nil, err
	}

	if !githubNameRe.MatchString(owner) || !githubNameRe.MatchString(repo) {
		return nil, fmt.Errorf("incorrect owner/repo: %q/%q", owner, repo)
	}

	return &Obj{name: owner + "/" + repo}, nil
}
