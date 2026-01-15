package gitlab

import (
	"fmt"
	"net/url"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

func (obj *Obj) getJSON(u string, out any) error {
	return lightweigit.GetJSON(obj, fmt.Sprintf("https://%s/api/v4/projects/%d/%s", obj.host, obj.id, u), &out)
}

// //

func (obj *Obj) Type() string {
	return "gitlab"
}

func (obj *Obj) Domain() string {
	return obj.host
}

func (obj *Obj) String() string {
	return obj.name
}

func (obj *Obj) URL() *url.URL {
	return lightweigit.BuildURL(
		"https",
		obj.host,
		obj.name,
		"",
	)
}
