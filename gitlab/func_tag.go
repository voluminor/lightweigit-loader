package gitlab

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
		"/-/tree/"+tag.name,
		"",
	)
}

func (tag *TagObj) ZIP() *url.URL {
	return lightweigit.BuildURL(
		"https",
		tag.Provider.host,
		fmt.Sprintf("api/v4/projects/%d/repository/archive.zip", tag.Provider.id),
		fmt.Sprintf("sha=%s", url.QueryEscape(tag.name)),
	)
}

func (tag *TagObj) TAR() *url.URL {
	return lightweigit.BuildURL(
		"https",
		tag.Provider.host,
		fmt.Sprintf("api/v4/projects/%d/repository/archive.tar.gz", tag.Provider.id),
		fmt.Sprintf("sha=%s", url.QueryEscape(tag.name)),
	)
}

// // // //

type tagItemObj struct {
	Name   string `json:"name"`
	Commit struct {
		ID string `json:"id"`
	} `json:"commit"`
}

// //

func (obj *Obj) TagLatest() (lightweigit.ProviderTagInterface, error) {
	var tags []tagItemObj
	if err := obj.getJSON("repository/tags?per_page=1&page=1&order_by=updated&sort=desc", &tags); err != nil {
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
	var t tagItemObj

	tagEsc := url.PathEscape(findTag)
	if err := obj.getJSON(fmt.Sprintf("repository/tags/%s", tagEsc), &t); err != nil {
		return nil, err
	}

	name := t.Name
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

	sent := 0
	for page := 1; ; page++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var tags []tagItemObj

		if err := obj.getJSON(fmt.Sprintf("repository/tags?per_page=%d&page=%d&order_by=updated&sort=desc", perPage, page), &tags); err != nil {
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
