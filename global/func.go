package global

import (
	"errors"
	"lightweigit"
	"lightweigit/bitbucket"
	"lightweigit/github"
	"lightweigit/gitlab"
	"lightweigit/gogsFamily"
)

// // // // // // // // // // // // // // // //

func Parse(raw string) (lightweigit.ProviderInterface, error) {
	if raw == "" {
		return nil, errors.New("an empty URL string")
	}

	gh, err := github.Parse(raw)
	if err == nil {
		return gh, nil
	}

	b, err := bitbucket.Parse(raw)
	if err == nil {
		return b, nil
	}

	gl, err := gitlab.Parse(raw)
	if err == nil {
		return gl, nil
	}

	return gogsFamily.Parse(raw)
}
