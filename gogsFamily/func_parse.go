package gogsFamily

import (
	"encoding/json"
	"errors"
	"fmt"
	"lightweigit"
	"net/url"
	"strings"
)

// // // // // // // // // // // // // // // //

func isNameOK(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func ownerRepoFromPathSegments(segments []string, i int) (owner string, repo string, ok bool) {
	if i < 0 || i+1 >= len(segments) {
		return "", "", false
	}
	owner = segments[i]
	repo = segments[i+1]
	repo = strings.TrimSuffix(repo, ".git")
	if !isNameOK(owner) || !isNameOK(repo) {
		return "", "", false
	}
	return owner, repo, true
}

func parseScpLikeAny(s string) (host string, name string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected SSH format: %q", s)
	}

	left := parts[0]
	at := strings.LastIndex(left, "@")
	if at < 0 || at == len(left)-1 {
		return "", "", fmt.Errorf("unexpected SSH format (no host): %q", s)
	}

	host = strings.ToLower(left[at+1:])
	if host == "" {
		return "", "", fmt.Errorf("unexpected SSH format (empty host): %q", s)
	}

	path := strings.Trim(parts[1], "/")
	if path == "" {
		return "", "", errors.New("there is no repository path in the url")
	}

	segs := strings.Split(path, "/")
	if len(segs) < 2 {
		return "", "", fmt.Errorf("expected a path of the type owner/repo, received: %q", path)
	}

	owner := segs[0]
	repo := strings.TrimSuffix(segs[1], ".git")
	if !isNameOK(owner) || !isNameOK(repo) {
		return "", "", fmt.Errorf("incorrect owner/repo: %q/%q", owner, repo)
	}

	return host, owner + "/" + repo, nil
}

// // // //

type versionObj struct {
	Version string `json:"version"`
}

func getJSONProbe(obj lightweigit.ProviderInterface, absURL string, out any) (int, error) {
	b, code, err := getBytes(obj, absURL, "application/json", 1<<20)
	if err != nil {
		return code, err
	}
	if out == nil {
		return code, nil
	}
	return code, json.Unmarshal(b, out)
}

func detectProvider(host string) (KindType, error) {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimRight(host, "/")
	if host == "" {
		return TypeUnknown, errors.New("empty host")
	}

	base := "https://" + host
	probe := &Obj{host: host, kind: TypeUnknown}
	var v versionObj

	code, err := getJSONProbe(probe, base+"/api/forgejo/v1/version", &v)
	if err == nil && strings.TrimSpace(v.Version) != "" {
		return TypeForgejo, nil
	}
	if err != nil && (code == 401 || code == 403) {
		return TypeForgejo, nil
	}

	code, err = getJSONProbe(probe, base+"/api/v1/version", &v)
	if err == nil && strings.TrimSpace(v.Version) != "" {
		if strings.Contains(v.Version, "+gitea-") {
			return TypeForgejo, nil
		}
		return TypeGitea, nil
	}
	if err != nil && (code == 401 || code == 403) {
		return TypeGitea, nil
	}

	return TypeUnknown, nil
}

func probeRepoAPI(host string, name string) bool {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimRight(host, "/")
	if host == "" || name == "" {
		return false
	}

	probe := &Obj{host: host, kind: TypeUnknown}

	u := "https://" + host + "/api/v1/repos/" + strings.TrimLeft(name, "/")
	_, code, err := getBytes(probe, u, "application/json", 256<<10)
	if err == nil {
		return true
	}

	if code == 401 || code == 403 {
		return true
	}
	if errors.Is(err, lightweigit.ErrNotFound) {
		return false
	}

	return false
}

// // // //

func Parse(raw string) (*Obj, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, errors.New("an empty URL string")
	}

	if !strings.Contains(s, "://") && strings.Contains(s, "@") && strings.Contains(s, ":") {
		host, name, err := parseScpLikeAny(s)
		if err != nil {
			return nil, err
		}

		kind, err := detectProvider(host)
		if err != nil {
			return nil, err
		}

		if kind == TypeUnknown {
			if probeRepoAPI(host, name) {
				kind = TypeGogs
			}
		}

		return &Obj{
			name: name,
			host: host,
			kind: kind,
		}, nil
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

	var segs []string
	for _, seg := range strings.Split(strings.Trim(u.Path, "/"), "/") {
		if seg == "" {
			continue
		}
		segs = append(segs, seg)
	}
	if len(segs) < 2 {
		return nil, fmt.Errorf("expected a path of the type /owner/repo, received: %q", u.Path)
	}

	cands := make([]string, 0)
	for i := 0; i <= len(segs)-2; i++ {
		owner, repo, ok := ownerRepoFromPathSegments(segs, i)
		if !ok {
			continue
		}
		cands = append(cands, owner+"/"+repo)
	}
	if len(cands) == 0 {
		return nil, fmt.Errorf("could not find owner/repo in path: %q", u.Path)
	}

	kind, err := detectProvider(u.Host)
	if err != nil {
		return nil, err
	}

	for _, c := range cands {
		if probeRepoAPI(u.Host, c) {
			if kind == TypeUnknown {
				kind = TypeGogs
			}
			return &Obj{
				name: c,
				host: u.Host,
				kind: kind,
			}, nil
		}
	}

	last := cands[len(cands)-1]
	return &Obj{
		name: last,
		host: u.Host,
		kind: kind,
	}, nil
}
