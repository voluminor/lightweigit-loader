package bitbucket

import (
	"context"
	"fmt"
	"net/url"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

func (tag *TagObj) Mod() target.ModType {
	return target.ModBitbucketTag
}

func (tag *TagObj) String() string {
	return tag.name
}

func (tag *TagObj) URL() *url.URL {
	return lightweigit.AddURL(
		tag.Provider.URL(),
		"/src/"+url.PathEscape(tag.name)+"/",
		"",
	)
}

func (tag *TagObj) ZIP() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"bitbucket.org",
		fmt.Sprintf("%s/get/%s.zip", tag.Provider.name, url.PathEscape(tag.name)),
		"",
	)
}

func (tag *TagObj) TAR() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"bitbucket.org",
		fmt.Sprintf("%s/get/%s.tar.gz", tag.Provider.name, url.PathEscape(tag.name)),
		"",
	)
}

// // // //

type tagItemObj struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
		Type string `json:"type"`
	} `json:"target"`
}

type tagsRespObj struct {
	Values   []tagItemObj `json:"values"`
	Next     string       `json:"next"`
	Page     int          `json:"page"`
	PageLen  int          `json:"pagelen"`
	Size     int          `json:"size"`
	Previous string       `json:"previous"`
}

// //

// TagLatest бере ПЕРШИЙ тег зі /refs/tags?pagelen=1&sort=-name (natural sorting по імені)
func (obj *Obj) TagLatest() (lightweigit.ProviderTagInterface, error) {
	var tr tagsRespObj
	if err := obj.getJSON("refs/tags?pagelen=1&sort=-name", &tr); err != nil {
		return nil, err
	}
	if len(tr.Values) == 0 {
		return nil, lightweigit.ErrNotFound
	}

	li := tr.Values[0]
	return &TagObj{
		Provider: obj,
		name:     li.Name,
	}, nil
}

func (obj *Obj) TagFind(findTag string) (lightweigit.ProviderTagInterface, error) {
	var ti tagItemObj
	if err := obj.getJSON(fmt.Sprintf("refs/tags/%s", url.PathEscape(findTag)), &ti); err != nil {
		return nil, err
	}

	name := ti.Name
	if name == "" {
		name = findTag
	}

	return &TagObj{
		Provider: obj,
		name:     name,
	}, nil
}

func (obj *Obj) TagsStream(ctx context.Context, out chan lightweigit.ProviderTagInterface, limit int) error {
	perPage := 100
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	u := fmt.Sprintf("refs/tags?pagelen=%d&sort=-name", perPage)
	sent := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var tr tagsRespObj
		if err := obj.getJSON(u, &tr); err != nil {
			return err
		}
		if len(tr.Values) == 0 {
			return nil
		}

		for _, li := range tr.Values {
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

		if tr.Next == "" {
			return nil
		}
		u = tr.Next
	}
}
