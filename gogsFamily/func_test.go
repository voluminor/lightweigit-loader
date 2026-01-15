package gogsFamily

import (
	"fmt"
	"testing"
)

// // // // // // // // // // // // // // // //

func TestName(t *testing.T) {
	for pos, url := range []string{
		"https://gitea.com/gitea/helm-gitea",
		"https://forgejo.skynet.ie/Computer_Society/minecraft_vanilla-enhanced-modpack",
		"https://forgejo.ellis.link/continuwuation/continuwuity",
		"https://gitea.champs-libres.be/Chill-project/manuals",
		"https://gitea.nouspiro.space/nou/tenmon",
		//	"https://gogs.am-networks.fr/am/ClearBrowserCaches",
		//	"https://git.cybergav.in/pythoncoder8/connect-4-game",
	} {
		t.Run(fmt.Sprintf("%d", pos), func(t *testing.T) {
			obj, err := Parse(url)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(obj.String(), obj.Kind().String())

			tag, err := obj.TagLatest()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(tag)

			rel, err := obj.ReleaseLatest()
			if err != nil {
				t.Fatal(err)
			}

			t.Log(rel.Name(), rel.URL())
			for _, ass := range rel.Assets() {
				t.Log(ass.Name(), ass.ContentType(), ass.Size(), ass.URL())
			}
		})
	}
}
