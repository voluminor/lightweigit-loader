package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/bitbucket"
	"github.com/voluminor/lightweigit-loader/github"
)

// // // // // // // // // // // // // // // //

// Mirrors the unexported success-body cap in lightweigit.GetJSON.
const jsonBodyCap = 8 << 20

// rewriteTransportObj redirects every request to the local test server,
// regardless of the hardcoded provider API host.
type rewriteTransportObj struct {
	host string
}

func (rt rewriteTransportObj) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.URL.Scheme = "http"
	r2.URL.Host = rt.host
	return http.DefaultTransport.RoundTrip(r2)
}

func swapHTTPClient(t *testing.T, srv *httptest.Server) {
	t.Helper()

	old := lightweigit.HttpClient
	lightweigit.HttpClient = &http.Client{
		Transport: rewriteTransportObj{host: strings.TrimPrefix(srv.URL, "http://")},
		Timeout:   30 * time.Second,
	}
	t.Cleanup(func() { lightweigit.HttpClient = old })
}

func githubObj(t *testing.T) *github.Obj {
	t.Helper()

	obj, err := github.Parse("https://github.com/owner/repo")
	if err != nil {
		t.Fatalf("github.Parse error: %v", err)
	}
	return obj
}

func collectReleases(t *testing.T, obj *github.Obj, limit int) ([]string, error) {
	t.Helper()

	out := make(chan lightweigit.ProviderReleaseInterface)
	errCh := make(chan error, 1)
	go func() {
		errCh <- obj.ReleasesStream(context.Background(), out, limit)
		close(out)
	}()

	var names []string
	for rel := range out {
		names = append(names, rel.Name())
	}
	return names, <-errCh
}

// //

func TestGetJSON_ResponseTooLarge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(strings.Repeat("x", jsonBodyCap+16)))
	}))
	defer srv.Close()

	var out map[string]any
	err := lightweigit.GetJSON(githubObj(t), srv.URL, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, lightweigit.ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got: %v", err)
	}
	if strings.Contains(err.Error(), "unexpected end of JSON") {
		t.Fatalf("decode error leaked instead of the sentinel: %v", err)
	}
}

func TestGetJSON_BodyAtCapStillDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body := "{\"a\":1}" + strings.Repeat(" ", jsonBodyCap-7)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	var out map[string]any
	if err := lightweigit.GetJSON(githubObj(t), srv.URL, &out); err != nil {
		t.Fatalf("expected success at exactly the cap, got: %v", err)
	}
	if out["a"] != float64(1) {
		t.Fatalf("unexpected decode result: %v", out)
	}
}

func TestReleasesStream_AdaptivePageShrink(t *testing.T) {
	const total = 30

	var (
		mu   sync.Mutex
		reqs [][2]int
	)

	// Page is oversized when its window holds >= 3 of the fat items r26..r30.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/releases") {
			http.NotFound(w, r)
			return
		}

		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		mu.Lock()
		reqs = append(reqs, [2]int{perPage, page})
		mu.Unlock()

		start := (page-1)*perPage + 1
		end := page * perPage
		if end > total {
			end = total
		}

		fatFrom, fatTo := start, end
		if fatFrom < 26 {
			fatFrom = 26
		}
		if fatTo-fatFrom+1 >= 3 {
			w.Write([]byte(strings.Repeat("x", jsonBodyCap+16)))
			return
		}

		items := make([]string, 0)
		for i := start; i <= end; i++ {
			items = append(items, fmt.Sprintf("{\"tag_name\":\"r%d\",\"name\":\"r%d\"}", i, i))
		}
		w.Write([]byte("[" + strings.Join(items, ",") + "]"))
	}))
	defer srv.Close()
	swapHTTPClient(t, srv)

	names, err := collectReleases(t, githubObj(t), 0)
	if err != nil {
		t.Fatalf("ReleasesStream error: %v", err)
	}

	want := make([]string, 0, total)
	for i := 1; i <= total; i++ {
		want = append(want, fmt.Sprintf("r%d", i))
	}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("item sequence mismatch (dups/gaps):\n got=%v\nwant=%v", names, want)
	}

	wantReqs := [][2]int{
		{50, 1}, // fat: window holds all 5 fat items
		{25, 1}, // ok: r1..r25
		{25, 2}, // fat: r26..r30
		{12, 3}, // fat: window 25..36
		{6, 5},  // fat: window 25..30
		{3, 9},  // ok: r25(skipped),r26,r27
		{3, 10}, // fat: r28..r30
		{1, 28}, // ok
		{1, 29}, // ok
		{1, 30}, // ok
		{1, 31}, // empty -> end of stream
	}
	mu.Lock()
	defer mu.Unlock()
	if fmt.Sprint(reqs) != fmt.Sprint(wantReqs) {
		t.Fatalf("request sequence mismatch:\n got=%v\nwant=%v", reqs, wantReqs)
	}
}

func TestReleasesStream_SingleItemOverCapFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(strings.Repeat("x", jsonBodyCap+16)))
	}))
	defer srv.Close()
	swapHTTPClient(t, srv)

	_, err := collectReleases(t, githubObj(t), 0)
	if !errors.Is(err, lightweigit.ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge after shrinking to 1, got: %v", err)
	}
}

func TestTagsStream_DefaultPerPage50(t *testing.T) {
	var (
		mu    sync.Mutex
		first string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if first == "" {
			first = r.URL.RawQuery
		}
		mu.Unlock()
		w.Write([]byte("[{\"name\":\"v1\"},{\"name\":\"v2\"}]"))
	}))
	defer srv.Close()
	swapHTTPClient(t, srv)

	out := make(chan lightweigit.ProviderTagInterface)
	errCh := make(chan error, 1)
	go func() {
		errCh <- githubObj(t).TagsStream(context.Background(), out, 0)
		close(out)
	}()

	var tags []string
	for tag := range out {
		tags = append(tags, tag.String())
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TagsStream error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("unexpected tags: %v", tags)
	}

	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(first, "per_page=50") {
		t.Fatalf("first request must use per_page=50, got query: %q", first)
	}
}

func TestStreamPages_InvalidParams(t *testing.T) {
	fetch := func(perPage, page int) ([]int, error) { return nil, nil }
	emit := func(int) error { return nil }

	if err := lightweigit.StreamPages(context.Background(), 0, 0, fetch, emit); err == nil {
		t.Fatal("expected error for perPage=0")
	}
	if err := lightweigit.StreamPages(context.Background(), -3, 10, fetch, emit); err == nil {
		t.Fatal("expected error for negative perPage")
	}
	if err := lightweigit.StreamPages[int](context.Background(), 50, 0, nil, emit); err == nil {
		t.Fatal("expected error for nil fetch")
	}
	if err := lightweigit.StreamPages(context.Background(), 50, 0, fetch, nil); err == nil {
		t.Fatal("expected error for nil emit")
	}
}

func TestReleasesStream_CancelUnblocksSend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`[{"tag_name":"r1","name":"r1"},{"tag_name":"r2","name":"r2"}]`))
	}))
	defer srv.Close()
	swapHTTPClient(t, srv)

	obj := githubObj(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan lightweigit.ProviderReleaseInterface)
	errCh := make(chan error, 1)
	go func() {
		errCh <- obj.ReleasesStream(ctx, out, 0)
	}()

	// Take the first item, then stop reading: the stream blocks on the
	// second send and only cancellation may release it.
	select {
	case <-out:
	case <-time.After(2 * time.Second):
		t.Fatal("stream produced no items")
	}
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stream is still blocked after cancel")
	}
}

func TestBitbucketTagsStream_FollowsNextCursor(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		if r.URL.Query().Get("page") == "2" {
			w.Write([]byte(`{"values":[{"name":"v3"}]}`))
			return
		}
		w.Write([]byte(`{"values":[{"name":"v1"},{"name":"v2"}],` +
			`"next":"https://api.bitbucket.org/2.0/repositories/owner/repo/refs/tags?pagelen=50&page=2"}`))
	}))
	defer srv.Close()
	swapHTTPClient(t, srv)

	obj, err := bitbucket.Parse("https://bitbucket.org/owner/repo")
	if err != nil {
		t.Fatalf("bitbucket.Parse error: %v", err)
	}

	out := make(chan lightweigit.ProviderTagInterface)
	errCh := make(chan error, 1)
	go func() {
		errCh <- obj.TagsStream(context.Background(), out, 0)
		close(out)
	}()

	var tags []string
	for tag := range out {
		tags = append(tags, tag.String())
	}
	if err := <-errCh; err != nil {
		t.Fatalf("TagsStream error: %v", err)
	}
	if strings.Join(tags, ",") != "v1,v2,v3" {
		t.Fatalf("unexpected tags: %v", tags)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 {
		t.Fatalf("expected 2 requests, got: %v", paths)
	}
	for _, p := range paths {
		if strings.Contains(p, "https:") {
			t.Fatalf("cursor URL got glued onto the API root: %q", p)
		}
	}
}
