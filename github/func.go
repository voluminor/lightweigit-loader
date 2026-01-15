package github

import (
	"fmt"
	"lightweigit"
	"net/url"
)

// // // // // // // // // // // // // // // //

func (obj *Obj) getJSON(u string, out any) error {
	return lightweigit.GetJSON(obj, fmt.Sprintf("https://api.github.com/repos/%s/%s", obj.name, u), &out)
}

// //

func (obj *Obj) Type() string {
	return "github"
}

func (obj *Obj) Domain() string {
	return "github.com"
}

func (obj *Obj) String() string {
	return obj.name
}

func (obj *Obj) URL() *url.URL {
	return lightweigit.BuildURL(
		"https",
		"github.com",
		obj.name,
		"",
	)
}
