// Code generated using '_generate/build_func'; DO NOT EDIT.
// Generation time: 2026-01-15T19:06:45Z

package global

import (
	"errors"
	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/bitbucket"
	"github.com/voluminor/lightweigit-loader/github"
	"github.com/voluminor/lightweigit-loader/gitlab"
	"github.com/voluminor/lightweigit-loader/gogsFamily"
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

	m0, err := bitbucket.UnmarshalTag(data)
	if err == nil {
		return m0, nil
	}

	m1, err := github.UnmarshalTag(data)
	if err == nil {
		return m1, nil
	}

	m2, err := gitlab.UnmarshalTag(data)
	if err == nil {
		return m2, nil
	}

	m3, err := gogsFamily.UnmarshalTag(data)
	if err == nil {
		return m3, nil
	}

	return nil, err
}

func UnmarshalRelease(data []byte) (lightweigit.ProviderReleaseInterface, error) {
	if len(data) < 5 {
		return nil, errors.New("not enough data")
	}

	m0, err := bitbucket.UnmarshalRelease(data)
	if err == nil {
		return m0, nil
	}

	m1, err := github.UnmarshalRelease(data)
	if err == nil {
		return m1, nil
	}

	m2, err := gitlab.UnmarshalRelease(data)
	if err == nil {
		return m2, nil
	}

	m3, err := gogsFamily.UnmarshalRelease(data)
	if err == nil {
		return m3, nil
	}

	return nil, err
}
