package github

import (
	"context"
	"fmt"
	"net/url"
	"path"

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
	return target.ModGithubRelease
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
		"/releases/tag/"+rel.tag.String(),
		"",
	)
}

func (rel *ReleaseObj) Tag() lightweigit.ProviderTagInterface {
	return rel.tag
}

func (rel *ReleaseObj) ZIP() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"api.github.com",
		fmt.Sprintf("repos/%s/zipball/%s", rel.Provider.name, rel.name),
		"",
	)
}

func (rel *ReleaseObj) TAR() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"api.github.com",
		fmt.Sprintf("repos/%s/tarball/%s", rel.Provider.name, rel.name),
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
	assets := make([]lightweigit.ProviderReleaseAssetInterface, 0, len(li.Assets))
	for _, a := range li.Assets {
		u, _ := url.Parse(a.BrowserDownloadURL)
		assets = append(assets, &ReleaseAssetObj{
			download:    *u,
			contentType: a.ContentType,
			size:        a.Size,
		})
	}

	return &ReleaseObj{
		Provider: obj,
		tag: &TagObj{
			Provider: obj,
			name:     li.TagName,
		},
		name:         li.Name,
		bodyMD:       li.Body,
		assets:       assets,
		isPrerelease: li.Prerelease,
	}
}

// //

func (obj *Obj) ReleaseLatest() (lightweigit.ProviderReleaseInterface, error) {
	var latest releaseItemObj
	if err := obj.getJSON("releases/latest", &latest); err == nil {
		ro := buildReleaseObj(obj, latest)
		if !ro.isPrerelease {
			return ro, nil
		}
	}

	perPage := 100
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?per_page=%d&page=%d", perPage, page), &rels); err != nil {
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
	var li releaseItemObj
	escaped := url.PathEscape(findRelease)
	err := obj.getJSON(fmt.Sprintf("releases/tags/%s", escaped), &li)
	if err == nil {
		return buildReleaseObj(obj, li), nil
	}
	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	perPage := 100
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?per_page=%d&page=%d", perPage, page), &rels); err != nil {
			return nil, err
		}
		if len(rels) == 0 {
			return nil, lightweigit.ErrNotFound
		}

		for _, li := range rels {
			if li.TagName == findRelease || li.Name == findRelease {
				return buildReleaseObj(obj, li), nil
			}
		}

		if len(rels) < perPage {
			return nil, lightweigit.ErrNotFound
		}
	}
}

func (obj *Obj) ReleasesStream(ctx context.Context, out chan lightweigit.ProviderReleaseInterface, limit int) error {
	perPage := 100
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	sent := 0
	for page := 1; ; page++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var rels []releaseItemObj
		if err := obj.getJSON(fmt.Sprintf("releases?per_page=%d&page=%d", perPage, page), &rels); err != nil {
			return err
		}
		if len(rels) == 0 {
			return nil
		}

		for _, li := range rels {
			if limit > 0 && sent >= limit {
				return nil
			}

			if ctx.Err() != nil {
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
