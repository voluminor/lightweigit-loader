package gogsFamily

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

func (tag *TagObj) String() string {
	return tag.name
}

func (tag *TagObj) URL() *url.URL {
	return lightweigit.AddURL(
		tag.Provider.URL(),
		"/src/tag/"+url.PathEscape(tag.name),
		"",
	)
}

func (tag *TagObj) ZIP() *url.URL {
	return lightweigit.AddURL(
		tag.Provider.URL(),
		"/archive/"+url.PathEscape(tag.name)+".zip",
		"",
	)
}

func (tag *TagObj) TAR() *url.URL {
	return lightweigit.AddURL(
		tag.Provider.URL(),
		"/archive/"+url.PathEscape(tag.name)+".tar.gz",
		"",
	)
}

// // // //

type tagItemObj struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Commit struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
}

// // // //

func (obj *Obj) TagLatest() (lightweigit.ProviderTagInterface, error) {
	var tags []tagItemObj
	err := obj.getJSON("tags?limit=1&page=1", &tags)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, lightweigit.ErrNotFound
	}
	return &TagObj{Provider: obj, name: tags[0].Name}, nil
}

func (obj *Obj) TagFind(findTag string) (lightweigit.ProviderTagInterface, error) {
	findTag = strings.TrimSpace(findTag)
	if findTag == "" {
		return nil, errors.New("empty tag")
	}

	var li tagItemObj
	err := obj.getJSON(fmt.Sprintf("tags/%s", url.PathEscape(findTag)), &li)
	if err != nil {
		return nil, err
	}

	name := li.Name
	if name == "" {
		name = findTag
	}
	return &TagObj{Provider: obj, name: name}, nil
}

func (obj *Obj) TagsStream(ctx context.Context, out chan lightweigit.ProviderTagInterface, limit int) error {
	perPage := 50
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	sent := 0
	for page := 1; ; page++ {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}

		var tags []tagItemObj
		err := obj.getJSON(fmt.Sprintf("tags?limit=%d&page=%d", perPage, page), &tags)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			return nil
		}

		for _, li := range tags {
			if limit > 0 && sent >= limit {
				return nil
			}
			if ctx != nil && ctx.Err() != nil {
				return ctx.Err()
			}

			out <- &TagObj{Provider: obj, name: li.Name}
			sent++
		}

		if len(tags) < perPage {
			return nil
		}
	}
}
