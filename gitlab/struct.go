package gitlab

import (
	"net/url"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

type Obj struct {
	name string
	host string

	id uint32
}

type TagObj struct {
	Provider *Obj
	name     string
}

type ReleaseAssetObj struct {
	download    url.URL
	contentType string
	size        uint32
}

type ReleaseObj struct {
	Provider     *Obj
	tag          lightweigit.ProviderTagInterface
	name         string
	bodyMD       string
	assets       []lightweigit.ProviderReleaseAssetInterface
	isPrerelease bool
}
