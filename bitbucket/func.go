package bitbucket

import (
	"fmt"
	"net/url"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

func (obj *Obj) getJSON(u string, out any) error {
	return lightweigit.GetJSON(obj, fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s", obj.name, u), &out)
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
