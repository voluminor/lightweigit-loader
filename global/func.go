package global

import (
	"errors"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/bitbucket"
	"github.com/voluminor/lightweigit-loader/github"
	"github.com/voluminor/lightweigit-loader/gitlab"
	"github.com/voluminor/lightweigit-loader/gogsFamily"
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
