package github

import (
	"errors"
	"testing"

	"github.com/voluminor/lightweigit-loader"
)

// // // // // // // // // // // // // // // //

// skipIfLimited turns provider-side blocking (rate limits, bot protection)
// into a skip: shared CI runner IPs are routinely throttled and that is not
// a code failure.
func skipIfLimited(t *testing.T, err error) {
	t.Helper()
	if errors.Is(err, lightweigit.ErrForbidden) || errors.Is(err, lightweigit.ErrTooManyRequests) {
		t.Skipf("provider blocked the request: %v", err)
	}
}

func TestName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in -short mode")
	}

	obj, err := Parse("https://github.com/AI-translate-book/template-EN-to-RU/commits/v0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(obj)

	tag, err := obj.TagLatest()
	if err != nil {
		skipIfLimited(t, err)
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
		skipIfLimited(t, err)
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
