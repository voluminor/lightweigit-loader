package gogsFamily

import (
	"net/url"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

type KindType byte

const (
	TypeUnknown KindType = iota
	TypeGitea
	TypeForgejo
	TypeGogs
)

func (k KindType) String() string {
	switch k {
	case TypeGitea:
		return "gitea"
	case TypeForgejo:
		return "forgejo"
	case TypeGogs:
		return "gogs"
	default:
		return "unknown"
	}
}

// // // //

type Obj struct {
	name string
	host string
	kind KindType
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
