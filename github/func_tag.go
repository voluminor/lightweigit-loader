package github

import (
	"context"
	"fmt"
	"lightweigit"
	"net/url"
)

// // // // // // // // // // // // // // // //

func (tag *TagObj) String() string {
	return tag.name
}

func (tag *TagObj) URL() *url.URL {
	return lightweigit.AddURL(
		tag.Provider.URL(),
		"/tag/"+tag.name,
		"",
	)
}

func (tag *TagObj) ZIP() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"github.com",
		fmt.Sprintf("%s/archive/refs/tags/%s.zip", tag.Provider.name, tag.name),
		"",
	)
}

func (tag *TagObj) TAR() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"github.com",
		fmt.Sprintf("%s/archive/refs/tags/%s.tar.gz", tag.Provider.name, tag.name),
		"",
	)
}

// // // //

type tagItemObj struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

type refRespObj struct {
	Ref    string `json:"ref"`
	URL    string `json:"url"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
		URL  string `json:"url"`
	} `json:"object"`
}

// //

func (obj *Obj) TagLatest() (lightweigit.ProviderTagInterface, error) {
	var tags []tagItemObj
	if err := obj.getJSON("tags?per_page=1&page=1", &tags); err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, lightweigit.ErrNotFound
	}

	li := tags[0]
	return &TagObj{
		Provider: obj,
		name:     li.Name,
	}, nil
}

func (obj *Obj) TagFind(findTag string) (lightweigit.ProviderTagInterface, error) {
	var rr refRespObj
	if err := obj.getJSON(fmt.Sprintf("git/ref/tags/%s", findTag), &rr); err != nil {
		return nil, err
	}

	return &TagObj{
		Provider: obj,
		name:     findTag,
	}, nil
}

func (obj *Obj) TagsStream(ctx context.Context, out chan lightweigit.ProviderTagInterface, limit int) error {
	perPage := 100
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	sent := 0
	for page := 1; ; page++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var tags []tagItemObj
		if err := obj.getJSON(fmt.Sprintf("tags?per_page=%d&page=%d", perPage, page), &tags); err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}

		for _, li := range tags {
			if limit > 0 && sent >= limit {
				return nil
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			out <- &TagObj{
				Provider: obj,
				name:     li.Name,
			}
			sent++
		}

		if len(tags) < perPage {
			return nil
		}
	}
}
