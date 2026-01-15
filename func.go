package lightweigit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/voluminor/lightweigit-loader/target"
)

// // // // // // // // // // // // // // // //

func UserAgent(obj ProviderInterface) string {
	return fmt.Sprintf("%s %s; %s (Goland %s %s)", target.GlobalName, target.GlobalVersion, obj.Type(), runtime.GOOS, runtime.GOARCH)
}

func GetJSON(obj ProviderInterface, u string, out any) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent(obj))
	req.Header.Set("Accept", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%s api error: %s: %s", obj.Type(), resp.Status, strings.TrimSpace(string(b)))
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// //

func BuildURL(scheme, host, path, query string) *url.URL {
	return &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawPath:  (&url.URL{Path: path}).EscapedPath(),
		OmitHost: host == "",
		RawQuery: query,
	}
}

func AddURL(u *url.URL, addToPath, setQuery string) *url.URL {
	u.Path += addToPath
	u.RawPath = (&url.URL{Path: u.Path}).EscapedPath()
	u.RawQuery = setQuery
	return u
}
