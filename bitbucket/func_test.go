package bitbucket

import "testing"

// // // // // // // // // // // // // // // //

func TestName(t *testing.T) {
	obj, err := Parse("https://bitbucket.org/bradleysmithllc/json-validator/src/master/")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(obj)

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
}
