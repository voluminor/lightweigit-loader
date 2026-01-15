package gitlab

import (
	"net/url"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

type byteObj struct {
	ID   uint32
	Name string
	Host string
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
			Host: tag.Provider.host,
			ID:   tag.Provider.id,
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
	if mod != target.ModGitlabTag {
		return nil, lightweigit.ErrModTag
	}

	return &TagObj{
		Provider: &Obj{
			name: dataObj.Obj.Name,
			host: dataObj.Obj.Host,
			id:   dataObj.Obj.ID,
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
			Name: rel.Provider.name,
			Host: rel.Provider.host,
			ID:   rel.Provider.id,
		},
		Tag: byteTagObj{
			Obj: byteObj{
				Name: rel.Provider.name,
				Host: rel.Provider.host,
				ID:   rel.Provider.id,
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
	if mod != target.ModGitlabRelease {
		return nil, lightweigit.ErrModTag
	}

	obj := &Obj{
		name: dataObj.Obj.Name,
		host: dataObj.Obj.Host,
		id:   dataObj.Obj.ID,
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
