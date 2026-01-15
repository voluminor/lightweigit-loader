package lightweigit

import (
	"context"
	"net/url"

	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

type ProviderTagInterface interface {
	Mod() target.ModType
	Marshal() []byte
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
	Mod() target.ModType
	Marshal() []byte
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
