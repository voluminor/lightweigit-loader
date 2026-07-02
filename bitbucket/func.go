package bitbucket

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

func (obj *Obj) getJSON(u string, out any) error {
	return lightweigit.GetJSON(obj, fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", obj.name, u), &out)
}

// getJSONAny follows Bitbucket cursor pagination: `next` is an absolute URL,
// everything else is relative to the repository API root.
func (obj *Obj) getJSONAny(u string, out any) error {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return lightweigit.GetJSON(obj, u, &out)
	}
	return obj.getJSON(u, out)
}

// //

func (obj *Obj) Type() string {
	return "bitbucket"
}

func (obj *Obj) Domain() string {
	return "bitbucket.org"
}

func (obj *Obj) String() string {
	return obj.name
}

func (obj *Obj) URL() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"bitbucket.org",
		obj.name,
		"",
	)
}
