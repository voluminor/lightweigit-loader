package global

import "testing"

// // // // // // // // // // // // // // // //

func TestGlobal(t *testing.T) {
	for _, url := range []string{
		"https://bitbucket.org/satt/or-pas/src/master/OrPasInstance.xsd",
		"https://gitea.com/gitea/util/src/branch/main/shellquote_test.go",
		"https://github.com/AI-translate-book/template-EN-to-RU/commits/v0.1.0",
		"https://gitlab.com/Meithal/cisson/-/blob/master/CMakeLists.txt?ref_type=heads",
	} {
		obj, err := Parse(url)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(obj.Type(), obj.String())
		}
	}
}
