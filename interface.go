package lightweigit

import (
	"context"
	"net/url"
)

// // // // // // // // // // // // // // // //

type ProviderTagInterface interface {
	String() string
	URL() *url.URL
	ZIP() *url.URL
	TAR() *url.URL
}

//

type ProviderReleaseAssetInterface interface {
	Name() string
	URL() *url.URL
	ContentType() string
	Size() uint32
}

type ProviderReleaseInterface interface {
	Name() string
	BodyMD() string
	URL() *url.URL
	Tag() ProviderTagInterface
	ZIP() *url.URL
	TAR() *url.URL
	Assets() []ProviderReleaseAssetInterface
	IsPrerelease() bool
}

// //

type ProviderInterface interface {
	Type() string
	Domain() string
	String() string
	URL() *url.URL

	TagLatest() (ProviderTagInterface, error)
	TagFind(string) (ProviderTagInterface, error)
	TagsStream(context.Context, chan ProviderTagInterface, int) error

	ReleaseLatest() (ProviderReleaseInterface, error)
	ReleaseFind(string) (ProviderReleaseInterface, error)
	ReleasesStream(context.Context, chan ProviderReleaseInterface, int) error
}
