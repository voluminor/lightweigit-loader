package gogsFamily

import (
	"fmt"
	"io"
	"lightweigit"
	"net/http"
	"net/url"
	"strings"
)

// // // // // // // // // // // // // // // //

func getBytes(obj lightweigit.ProviderInterface, absURL string, accept string, limitBytes int64) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodGet, absURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", lightweigit.UserAgent(obj))
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := lightweigit.HttpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, resp.StatusCode, lightweigit.ErrNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		t := "unknown"
		if obj != nil {
			t = obj.Type()
		}
		return nil, resp.StatusCode, fmt.Errorf("%s api error: %s: %s", t, resp.Status, strings.TrimSpace(string(b)))
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, limitBytes))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}

func (obj *Obj) getJSON(u string, out any) error {
	return lightweigit.GetJSON(obj, fmt.Sprintf("https://%s/api/v1/repos/%s/%s", obj.host, obj.name, u), out)
}

// // // // // // // // // // // // // // // //

func (obj *Obj) Kind() KindType { return obj.kind }

func (obj *Obj) Type() string {
	return obj.kind.String()
}

func (obj *Obj) Domain() string {
	return obj.host
}

func (obj *Obj) String() string {
	return obj.name
}

func (obj *Obj) URL() *url.URL {
	if obj == nil {
		return nil
	}
	return lightweigit.BuildURL(
		"https",
		obj.host,
		obj.name,
		"",
	)
}
