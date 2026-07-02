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

// TagLatest бере ПЕРШИЙ тег зі /refs/tags?pagelen=1&sort=-target.date (сортування за датою коміта)
func (obj *Obj) TagLatest() (lightweigit.ProviderTagInterface, error) {
	var tr tagsRespObj
	if err := obj.getJSON("refs/tags?pagelen=1&sort=-target.date", &tr); err != nil {
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

// No adaptive page shrink here: Bitbucket paginates via opaque `next` cursor
// URLs with pagelen baked in, so a mid-stream window remap is not possible.
// The 50 default plus the 8 MiB GetJSON cap keeps pages within bounds.
func (obj *Obj) TagsStream(ctx context.Context, out chan lightweigit.ProviderTagInterface, limit int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	perPage := 50
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	u := fmt.Sprintf("refs/tags?pagelen=%d&sort=-target.date", perPage)
	sent := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var tr tagsRespObj
		if err := obj.getJSONAny(u, &tr); err != nil {
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

			if err := lightweigit.Send[lightweigit.ProviderTagInterface](ctx, out, &TagObj{
				Provider: obj,
				name:     li.Name,
			}); err != nil {
				return err
			}
			sent++
		}

		if tr.Next == "" {
			return nil
		}
		u = tr.Next
	}
}
