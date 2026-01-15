package github

import (
	"lightweigit"
	"net/url"
)

// // // // // // // // // // // // // // // //

type Obj struct {
	name string
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
