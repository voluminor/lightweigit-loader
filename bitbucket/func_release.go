package bitbucket

import (
	"context"
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
	return target.ModBitbucketRelease
}

func (rel *ReleaseObj) Name() string {
	return rel.name
}

func (rel *ReleaseObj) BodyMD() string {
	return rel.bodyMD
}

func (rel *ReleaseObj) URL() *url.URL {
	return rel.tag.URL()
}

func (rel *ReleaseObj) Tag() lightweigit.ProviderTagInterface {
	return rel.tag
}

func (rel *ReleaseObj) ZIP() *url.URL {
	return rel.tag.ZIP()
}

func (rel *ReleaseObj) TAR() *url.URL {
	return rel.tag.TAR()
}

func (rel *ReleaseObj) Assets() []lightweigit.ProviderReleaseAssetInterface {
	return rel.assets
}

func (rel *ReleaseObj) IsPrerelease() bool {
	return rel.isPrerelease
}

// // // //

type downloadsLinkObj struct {
	Href string `json:"href"`
}

type downloadItemObj struct {
	Name  string `json:"name"`
	Size  uint32 `json:"size"`
	Links struct {
		Self downloadsLinkObj `json:"self"`
	} `json:"links"`
}

type downloadsRespObj struct {
	Values   []downloadItemObj `json:"values"`
	Next     string            `json:"next"`
	Page     int               `json:"page"`
	PageLen  int               `json:"pagelen"`
	Size     int               `json:"size"`
	Previous string            `json:"previous"`
}

func (obj *Obj) getJSONAny(u string, out any) error {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return lightweigit.GetJSON(obj, u, &out)
	}
	return obj.getJSON(u, out)
}

func (obj *Obj) buildRelease(tagName string, assets []lightweigit.ProviderReleaseAssetInterface) *ReleaseObj {
	return &ReleaseObj{
		Provider: obj,
		tag: &TagObj{
			Provider: obj,
			name:     tagName,
		},
		name:         tagName,
		bodyMD:       "",
		assets:       assets,
		isPrerelease: false,
	}
}

func (obj *Obj) listDownloads(ctx context.Context, limit int) ([]lightweigit.ProviderReleaseAssetInterface, error) {
	perPage := 100
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	assets := make([]lightweigit.ProviderReleaseAssetInterface, 0)

	u := fmt.Sprintf("downloads?pagelen=%d", perPage)
	sent := 0
	for {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var dr downloadsRespObj
		if err := obj.getJSONAny(u, &dr); err != nil {
			return nil, err
		}
		if len(dr.Values) == 0 {
			return assets, nil
		}

		for _, it := range dr.Values {
			if limit > 0 && sent >= limit {
				return assets, nil
			}
			if ctx != nil && ctx.Err() != nil {
				return nil, ctx.Err()
			}

			dlURL := it.Links.Self.Href
			if dlURL == "" {
				dlURL = fmt.Sprintf(
					"https://api.bitbucket.org/2.0/repositories/%s/downloads/%s",
					obj.name,
					url.PathEscape(it.Name),
				)
			}

			parsed, _ := url.Parse(dlURL)
			assets = append(assets, &ReleaseAssetObj{
				download:    *parsed,
				contentType: "",
				size:        it.Size,
			})
			sent++
		}

		if dr.Next == "" {
			return assets, nil
		}
		u = dr.Next
	}
}

// //

func (obj *Obj) ReleaseLatest() (lightweigit.ProviderReleaseInterface, error) {
	t, err := obj.TagLatest()
	if err != nil {
		return nil, err
	}

	assets, err := obj.listDownloads(context.Background(), 0)
	if err != nil {
		assets = nil
	}

	return obj.buildRelease(t.String(), assets), nil
}

func (obj *Obj) ReleaseFind(findRelease string) (lightweigit.ProviderReleaseInterface, error) {
	t, err := obj.TagFind(findRelease)
	if err != nil {
		return nil, err
	}

	assets, err := obj.listDownloads(context.Background(), 0)
	if err != nil {
		assets = nil
	}

	return obj.buildRelease(t.String(), assets), nil
}

func (obj *Obj) ReleasesStream(ctx context.Context, out chan lightweigit.ProviderReleaseInterface, limit int) error {
	perPage := 100
	if limit > 0 && limit < perPage {
		perPage = limit
	}

	assets, err := obj.listDownloads(ctx, 0)
	if err != nil {
		assets = nil
	}

	u := fmt.Sprintf("refs/tags?pagelen=%d&sort=-name", perPage)
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

			out <- obj.buildRelease(li.Name, assets)
			sent++
		}

		if tr.Next == "" {
			return nil
		}
		u = tr.Next
	}
}
