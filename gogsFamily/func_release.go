package gogsFamily

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

func (a *ReleaseAssetObj) Name() string {
	return path.Base(a.download.Path)
}

func (a *ReleaseAssetObj) URL() *url.URL {
	return &a.download
}

func (a *ReleaseAssetObj) ContentType() string {
	return a.contentType
}

func (a *ReleaseAssetObj) Size() uint32 {
	return a.size
}

//

func (rel *ReleaseObj) Mod() target.ModType {
	return target.ModGogsFamilyRelease
}

func (rel *ReleaseObj) Name() string {
	return rel.name
}

func (rel *ReleaseObj) BodyMD() string {
	return rel.bodyMD
}

func (rel *ReleaseObj) URL() *url.URL {
	return lightweigit.AddURL(
		rel.Provider.URL(),
		"/releases/tag/"+url.PathEscape(rel.tag.String()),
		"",
	)
}

func (rel *ReleaseObj) Tag() lightweigit.ProviderTagInterface {
	return rel.tag
}

func (rel *ReleaseObj) ZIP() *url.URL {
	return lightweigit.AddURL(
		rel.Provider.URL(),
		"/archive/"+url.PathEscape(rel.tag.String())+".zip",
		"",
	)
}

func (rel *ReleaseObj) TAR() *url.URL {
	return lightweigit.AddURL(
		rel.Provider.URL(),
		"/archive/"+url.PathEscape(rel.tag.String())+".tar.gz",
		"",
	)
}

func (rel *ReleaseObj) Assets() []lightweigit.ProviderReleaseAssetInterface {
	return rel.assets
}

func (rel *ReleaseObj) IsPrerelease() bool {
	return rel.isPrerelease
}

// // // //

type releaseAssetItemObj struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
	Size               uint32 `json:"size"`
}

type releaseItemObj struct {
	TagName    string                `json:"tag_name"`
	Name       string                `json:"name"`
	Body       string                `json:"body"`
	Draft      bool                  `json:"draft"`
	Prerelease bool                  `json:"prerelease"`
	Assets     []releaseAssetItemObj `json:"assets"`
}

func buildReleaseObj(obj *Obj, li releaseItemObj) *ReleaseObj {
	name := strings.TrimSpace(li.Name)
	if name == "" {
		name = strings.TrimSpace(li.TagName)
	}

	tagName := strings.TrimSpace(li.TagName)
	if tagName == "" {
		tagName = name
	}

	assets := make([]lightweigit.ProviderReleaseAssetInterface, 0, len(li.Assets))
	for _, a := range li.Assets {
		u, err := url.Parse(a.BrowserDownloadURL)
		if err != nil || u == nil {
			continue
		}

		assets = append(assets, &ReleaseAssetObj{
			download:    *u,
			contentType: a.ContentType,
			size:        a.Size,
		})
	}

	isPrerelease := li.Prerelease || li.Draft

	return &ReleaseObj{
		Provider: obj,
		tag: &TagObj{
			Provider: obj,
			name:     tagName,
		},
		name:         name,
		bodyMD:       li.Body,
		assets:       assets,
		isPrerelease: isPrerelease,
	}
}

// //

func (obj *Obj) ReleaseLatest() (lightweigit.ProviderReleaseInterface, error) {
	var latest releaseItemObj
	err := obj.getJSON("releases/latest", &latest)
	if err == nil {
		ro := buildReleaseObj(obj, latest)
		if !ro.isPrerelease {
			return ro, nil
		}
	}
	if err != nil && !errors.Is(err, lightweigit.ErrNotFound) {
		return nil, err
	}

	perPage := 50
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?limit=%d&page=%d", perPage, page), &rels); err != nil {
			return nil, err
		}
		if len(rels) == 0 {
			return nil, lightweigit.ErrNotFound
		}

		for _, li := range rels {
			ro := buildReleaseObj(obj, li)
			if !ro.isPrerelease {
				return ro, nil
			}
		}

		if len(rels) < perPage {
			return nil, lightweigit.ErrNotFound
		}
	}
}

func (obj *Obj) ReleaseFind(findRelease string) (lightweigit.ProviderReleaseInterface, error) {
	findRelease = strings.TrimSpace(findRelease)
	if findRelease == "" {
		return nil, errors.New("empty release")
	}

	var li releaseItemObj
	err := obj.getJSON(fmt.Sprintf("releases/tags/%s", url.PathEscape(findRelease)), &li)
	if err == nil {
		return buildReleaseObj(obj, li), nil
	}
	if err != nil && !errors.Is(err, lightweigit.ErrNotFound) {
		return nil, err
	}

	perPage := 50
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?limit=%d&page=%d", perPage, page), &rels); err != nil {
			return nil, err
		}
		if len(rels) == 0 {
			return nil, lightweigit.ErrNotFound
		}

		for _, r := range rels {
			if r.TagName == findRelease || r.Name == findRelease {
				return buildReleaseObj(obj, r), nil
			}
		}

		if len(rels) < perPage {
			return nil, lightweigit.ErrNotFound
		}
	}
}

func (obj *Obj) ReleasesStream(ctx context.Context, out chan lightweigit.ProviderReleaseInterface, limit int) error {
	perPage := 50
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	sent := 0
	for page := 1; ; page++ {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}

		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?limit=%d&page=%d", perPage, page), &rels); err != nil {
			return err
		}
		if len(rels) == 0 {
			return nil
		}

		for _, li := range rels {
			if limit > 0 && sent >= limit {
				return nil
			}
			if ctx != nil && ctx.Err() != nil {
				return ctx.Err()
			}

			out <- buildReleaseObj(obj, li)
			sent++
		}

		if len(rels) < perPage {
			return nil
		}
	}
}
