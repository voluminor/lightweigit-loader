// Code generated using '_generate/build_func'; DO NOT EDIT.
// Generation time: 2026-07-02T17:03:34Z

package global

import (
	"errors"
	"fmt"
	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/bitbucket"
	"github.com/voluminor/lightweigit-loader/github"
	"github.com/voluminor/lightweigit-loader/gitlab"
	"github.com/voluminor/lightweigit-loader/gogsFamily"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // //

func Parse(raw string) (lightweigit.ProviderInterface, error) {
	if raw == "" {
		return nil, errors.New("an empty URL string")
	}

	m0, err := bitbucket.Parse(raw)
	if err == nil {
		return m0, nil
	}

	m1, err := github.Parse(raw)
	if err == nil {
		return m1, nil
	}

	m2, err := gitlab.Parse(raw)
	if err == nil {
		return m2, nil
	}

	m3, err := gogsFamily.Parse(raw)
	if err == nil {
		return m3, nil
	}

	return nil, err
}

//

func UnmarshalTag(data []byte) (lightweigit.ProviderTagInterface, error) {
	if len(data) < 5 {
		return nil, errors.New("not enough data")
	}

	switch target.ModType(data[0]) {
	case target.ModBitbucketTag:
		return bitbucket.UnmarshalTag(data)
	case target.ModGithubTag:
		return github.UnmarshalTag(data)
	case target.ModGitlabTag:
		return gitlab.UnmarshalTag(data)
	case target.ModGogsFamilyTag:
		return gogsFamily.UnmarshalTag(data)
	default:
		return nil, fmt.Errorf("unknown tag mod type: %d", data[0])
	}
}

func UnmarshalRelease(data []byte) (lightweigit.ProviderReleaseInterface, error) {
	if len(data) < 5 {
		return nil, errors.New("not enough data")
	}

	switch target.ModType(data[0]) {
	case target.ModBitbucketRelease:
		return bitbucket.UnmarshalRelease(data)
	case target.ModGithubRelease:
		return github.UnmarshalRelease(data)
	case target.ModGitlabRelease:
		return gitlab.UnmarshalRelease(data)
	case target.ModGogsFamilyRelease:
		return gogsFamily.UnmarshalRelease(data)
	default:
		return nil, fmt.Errorf("unknown release mod type: %d", data[0])
	}
}
