package bitbucket

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// // // // // // // // // // // // // // // //

var bbNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func parseScpLikeBitbucket(s string) (string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected SSH format: %q", s)
	}

	left := parts[0]
	if !strings.HasSuffix(strings.ToLower(left), "@bitbucket.org") {
		return "", fmt.Errorf("not Bitbucket SSH address: %q", s)
	}

	path := parts[1]
	workspace, repo, err := workspaceRepoFromPath(path)
	if err != nil {
		return "", err
	}

	return workspace + "/" + repo, nil
}

func workspaceRepoFromPath(p string) (string, string, error) {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return "", "", errors.New("there is no repository path in the url")
	}

	segments := strings.Split(trimmed, "/")
	if len(segments) < 2 {
		return "", "", fmt.Errorf("expected a path of the type /workspace/repo, received: %q", p)
	}

	workspace := segments[0]
	repo := segments[1]
	repo = strings.TrimSuffix(repo, ".git")

	if workspace == "" || repo == "" {
		return "", "", fmt.Errorf("workspace or repo is empty in the path: %q", p)
	}

	if !bbNameRe.MatchString(workspace) {
		return "", "", fmt.Errorf("incorrect workspace: %q", workspace)
	}
	if !bbNameRe.MatchString(repo) {
		return "", "", fmt.Errorf("incorrect repo: %q", repo)
	}

	return workspace, repo, nil
}

// // // //

func Parse(raw string) (*Obj, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, errors.New("an empty URL string")
	}

	if strings.Contains(s, "@bitbucket.org:") && !strings.Contains(s, "://") {
		name, err := parseScpLikeBitbucket(s)
		if err != nil {
			return nil, err
		}
		return &Obj{name: name}, nil
	}

	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("URL could not be parsed: %w", err)
	}

	host := strings.ToLower(u.Hostname())
	if host == "www.bitbucket.org" {
		host = "bitbucket.org"
	}
	if host != "bitbucket.org" {
		return nil, fmt.Errorf("not Bitbucket host: %q", host)
	}

	workspace, repo, err := workspaceRepoFromPath(u.Path)
	if err != nil {
		return nil, err
	}

	return &Obj{name: workspace + "/" + repo}, nil
}
