package gitlab

import (
	"context"
	"fmt"
	"lightweigit"
	"net/url"
	"path"
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

func (rel *ReleaseObj) Name() string {
	return rel.name
}

func (rel *ReleaseObj) BodyMD() string {
	return rel.bodyMD
}

func (rel *ReleaseObj) URL() *url.URL {
	return lightweigit.AddURL(
		rel.Provider.URL(),
		"/-/releases/"+url.PathEscape(rel.tag.String()),
		"",
	)
}

func (rel *ReleaseObj) Tag() lightweigit.ProviderTagInterface {
	return rel.tag
}

func (rel *ReleaseObj) ZIP() *url.URL {
	return lightweigit.BuildURL(
		"https",
		rel.Provider.host,
		fmt.Sprintf("api/v4/projects/%d/repository/archive.zip", rel.Provider.id),
		fmt.Sprintf("sha=%s", url.QueryEscape(rel.tag.String())),
	)
}

func (rel *ReleaseObj) TAR() *url.URL {
	return lightweigit.BuildURL(
		"https",
		rel.Provider.host,
		fmt.Sprintf("api/v4/projects/%d/repository/archive.tar.gz", rel.Provider.id),
		fmt.Sprintf("sha=%s", url.QueryEscape(rel.tag.String())),
	)
}

func (rel *ReleaseObj) Assets() []lightweigit.ProviderReleaseAssetInterface {
	return rel.assets
}

func (rel *ReleaseObj) IsPrerelease() bool {
	return rel.isPrerelease
}

// // // //

type releaseLinkItemObj struct {
	ID             uint32 `json:"id"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	DirectAssetURL string `json:"direct_asset_url"`
	LinkType       string `json:"link_type"`
}

type releaseItemObj struct {
	TagName           string `json:"tag_name"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	UpcomingRelease   bool   `json:"upcoming_release"`
	HistoricalRelease bool   `json:"historical_release"`
	Assets            struct {
		Links []releaseLinkItemObj `json:"links"`
	} `json:"assets"`
}

func buildReleaseObj(obj *Obj, li releaseItemObj) *ReleaseObj {
	assets := make([]lightweigit.ProviderReleaseAssetInterface, 0, len(li.Assets.Links))
	for _, a := range li.Assets.Links {
		uStr := a.DirectAssetURL
		if uStr == "" {
			uStr = a.URL
		}
		u, err := url.Parse(uStr)
		if err != nil || u == nil {
			continue
		}

		assets = append(assets, &ReleaseAssetObj{
			download:    *u,
			contentType: a.LinkType,
		})
	}

	name := li.Name
	if name == "" {
		name = li.TagName
	}

	return &ReleaseObj{
		Provider: obj,
		tag: &TagObj{
			Provider: obj,
			name:     li.TagName,
		},
		name:         name,
		bodyMD:       li.Description,
		assets:       assets,
		isPrerelease: li.UpcomingRelease,
	}
}

// //

func (obj *Obj) ReleaseLatest() (lightweigit.ProviderReleaseInterface, error) {
	var latest releaseItemObj
	if err := obj.getJSON("releases/permalink/latest", &latest); err == nil {
		ro := buildReleaseObj(obj, latest)
		if !ro.isPrerelease {
			return ro, nil
		}
	}

	perPage := 100
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(
			fmt.Sprintf("releases?per_page=%d&page=%d&order_by=released_at&sort=desc", perPage, page),
			&rels,
		); err != nil {
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
	err := obj.getJSON(fmt.Sprintf("releases/%s", escaped), &li)
	if err == nil {
		return buildReleaseObj(obj, li), nil
	}
	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	perPage := 100
	for page := 1; ; page++ {
		var rels []releaseItemObj
		if err := obj.getJSON(
			fmt.Sprintf("releases?per_page=%d&page=%d&order_by=released_at&sort=desc", perPage, page),
			&rels,
		); err != nil {
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
		if err := obj.getJSON(
			fmt.Sprintf("releases?per_page=%d&page=%d&order_by=released_at&sort=desc", perPage, page),
			&rels,
		); err != nil {
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
