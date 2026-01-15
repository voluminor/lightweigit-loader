package github

import "testing"

// // // // // // // // // // // // // // // //

func TestName(t *testing.T) {
	obj, err := Parse("https://github.com/AI-translate-book/template-EN-to-RU/commits/v0.1.0")
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
