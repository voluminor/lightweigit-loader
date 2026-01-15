package gitlab

import "testing"

// // // // // // // // // // // // // // // //

func TestName(t *testing.T) {
	obj, err := Parse("https://gitlab.com/gitlab-org/gitlab-foss/-/tags")
	if err != nil {
		t.Fatal(err)
	}

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
