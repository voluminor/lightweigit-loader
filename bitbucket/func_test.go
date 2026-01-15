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

	data := tag.Marshal()
	bTag, err := UnmarshalTag(data)
	if err != nil {
		t.Fatal(err)
	}
	if bTag.URL().String() != tag.URL().String() {
		t.Fatal("tag does not match:", bTag.URL(), tag.URL())
	}

	rel, err := obj.ReleaseLatest()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rel.Name(), rel.URL())
	for _, ass := range rel.Assets() {
		t.Log(ass.Name(), ass.ContentType(), ass.Size(), ass.URL())
	}

	data = rel.Marshal()
	bRel, err := UnmarshalRelease(data)
	if err != nil {
		t.Fatal(err)
	}
	if bRel.URL().String() != rel.URL().String() {
		t.Fatal("release does not match:", bRel.URL(), rel.URL())
	}
}
