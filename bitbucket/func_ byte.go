package bitbucket

import (
	"net/url"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

type byteObj struct {
	Name string
}
type byteTagObj struct {
	Obj  byteObj
	Name string
}

//

func (tag *TagObj) Marshal() []byte {
	dataObj := byteTagObj{
		Obj: byteObj{
			Name: tag.Provider.name,
		},
		Name: tag.name,
	}
	return lightweigit.Marshal(tag.Mod(), dataObj)
}

func UnmarshalTag(data []byte) (lightweigit.ProviderTagInterface, error) {
	dataObj := new(byteTagObj)
	mod, err := lightweigit.Unmarshal(data, dataObj)
	if err != nil {
		return nil, err
	}
	if mod != target.ModBitbucketTag {
		return nil, lightweigit.ErrModTag
	}

	return &TagObj{
		Provider: &Obj{
			name: dataObj.Obj.Name,
		},
		name: dataObj.Name,
	}, nil
}

// // // //

type byteAssetObj struct {
	Size        uint32
	ContentType string
	DownloadURL string
}
type byteReleaseObj struct {
	Obj          byteObj
	Tag          byteTagObj
	Name         string
	BodyMD       string
	IsPrerelease bool
	Assets       []byteAssetObj
}

//

func (rel *ReleaseObj) Marshal() []byte {
	dataObj := byteReleaseObj{
		Obj: byteObj{
			rel.Provider.name,
		},
		Tag: byteTagObj{
			Obj: byteObj{
				rel.Provider.name,
			},
			Name: rel.tag.String(),
		},
		Name:         rel.name,
		BodyMD:       rel.bodyMD,
		IsPrerelease: rel.isPrerelease,
		Assets:       make([]byteAssetObj, 0),
	}
	for _, asset := range rel.assets {
		dataObj.Assets = append(dataObj.Assets, byteAssetObj{
			Size:        asset.Size(),
			ContentType: asset.ContentType(),
			DownloadURL: asset.URL().String(),
		})
	}

	return lightweigit.Marshal(rel.Mod(), dataObj)
}

func UnmarshalRelease(data []byte) (lightweigit.ProviderReleaseInterface, error) {
	dataObj := new(byteReleaseObj)
	mod, err := lightweigit.Unmarshal(data, dataObj)
	if err != nil {
		return nil, err
	}
	if mod != target.ModBitbucketRelease {
		return nil, lightweigit.ErrModTag
	}

	obj := &Obj{
		name: dataObj.Obj.Name,
	}
	tag := &TagObj{
		Provider: obj,
		name:     dataObj.Tag.Name,
	}
	release := &ReleaseObj{
		Provider:     obj,
		tag:          tag,
		name:         dataObj.Name,
		bodyMD:       dataObj.BodyMD,
		isPrerelease: dataObj.IsPrerelease,
		assets:       make([]lightweigit.ProviderReleaseAssetInterface, 0),
	}
	for _, asset := range dataObj.Assets {
		u, _ := url.Parse(asset.DownloadURL)
		release.assets = append(release.assets, &ReleaseAssetObj{
			size:        asset.Size,
			contentType: asset.ContentType,
			download:    *u,
		})
	}

	return release, nil
}
