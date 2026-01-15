package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

var gitlabNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func parseScpLikeGitLab(s string) (host string, name string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected SSH format: %q", s)
	}

	left := parts[0]
	idx := strings.LastIndex(left, "@")
	if idx < 0 || idx == len(left)-1 {
		return "", "", fmt.Errorf("unexpected SSH format: %q", s)
	}

	host = strings.ToLower(left[idx+1:])
	if host == "" {
		return "", "", fmt.Errorf("empty host in SSH address: %q", s)
	}

	path := parts[1]
	namespace, repo, err := namespaceRepoFromPath(path)
	if err != nil {
		return "", "", err
	}

	return host, namespace + "/" + repo, nil
}

func namespaceRepoFromPath(p string) (string, string, error) {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return "", "", errors.New("there is no repository path in the url")
	}

	segments := strings.Split(trimmed, "/")
	for i, seg := range segments {
		if seg == "-" {
			segments = segments[:i]
			break
		}
	}

	if len(segments) < 2 {
		return "", "", fmt.Errorf("expected a path of the type /namespace/repo, received: %q", p)
	}

	repo := strings.TrimSuffix(segments[len(segments)-1], ".git")
	namespaceParts := segments[:len(segments)-1]

	if repo == "" || len(namespaceParts) == 0 {
		return "", "", fmt.Errorf("namespace or repo is empty in the path: %q", p)
	}

	for _, part := range namespaceParts {
		if part == "" || !gitlabNameRe.MatchString(part) {
			return "", "", fmt.Errorf("incorrect namespace segment: %q", part)
		}
	}
	if !gitlabNameRe.MatchString(repo) {
		return "", "", fmt.Errorf("incorrect repo: %q", repo)
	}

	return strings.Join(namespaceParts, "/"), repo, nil
}

func validateGitLab(host, name string) (*Obj, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/api/v4/projects/%s", host, url.PathEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gitlab check metadata")
	req.Header.Set("Accept", "application/json")

	resp, err := lightweigit.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var md struct {
			Id uint32 `json:"id"`
		}
		if err := json.Unmarshal(body, &md); err != nil {
			return nil, fmt.Errorf("metadata is not JSON: %w", err)
		}
		return &Obj{host: host, name: name, id: md.Id}, nil
	}

	return nil, fmt.Errorf("metadata endpoint returned %s", resp.Status)
}

// // // //

func Parse(raw string) (*Obj, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, errors.New("an empty URL string")
	}

	if strings.Contains(s, "@") && strings.Contains(s, ":") && !strings.Contains(s, "://") {
		host, name, err := parseScpLikeGitLab(s)
		if err != nil {
			return nil, err
		}
		return validateGitLab(host, name)
	}

	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("URL could not be parsed: %w", err)
	}

	if u.Host == "" {
		return nil, fmt.Errorf("URL has no host: %q", raw)
	}

	host := strings.ToLower(u.Host)
	if strings.HasPrefix(host, "www.gitlab.com") {
		host = strings.Replace(host, "www.gitlab.com", "gitlab.com", 1)
	}

	namespace, repo, err := namespaceRepoFromPath(u.Path)
	if err != nil {
		return nil, err
	}

	return validateGitLab(host, namespace+"/"+repo)
}
