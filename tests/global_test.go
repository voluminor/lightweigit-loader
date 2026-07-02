package tests

import (
	"errors"
	"strings"
	"testing"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/target/global"
)

// // // // // // // // // // // // // // // //

func TestGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in -short mode")
	}

	for _, url := range []string{
		"https://bitbucket.org/satt/or-pas/src/master/OrPasInstance.xsd",
		"https://gitea.com/gitea/util/src/branch/main/shellquote_test.go",
		"https://github.com/AI-translate-book/template-EN-to-RU/commits/v0.1.0",
		"https://gitlab.com/Meithal/cisson/-/blob/master/CMakeLists.txt?ref_type=heads",
	} {
		obj, err := global.Parse(url)
		if err != nil {
			// global.Parse returns only the last provider's error, so the
			// rate-limit sentinel may be lost along the chain; fall back to
			// matching the status text.
			msg := err.Error()
			if errors.Is(err, lightweigit.ErrForbidden) || errors.Is(err, lightweigit.ErrTooManyRequests) ||
				strings.Contains(msg, "403") || strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") {
				t.Skipf("provider blocked the request: %v", err)
			}
			t.Error(err)
		} else {
			t.Log(obj.Type(), obj.String())
		}
	}
}
