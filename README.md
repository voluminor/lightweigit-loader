[![Go Report Card](https://goreportcard.com/badge/github.com/voluminor/lightweigit-loader)](https://goreportcard.com/report/github.com/voluminor/lightweigit-loader)

![GitHub repo file or directory count](https://img.shields.io/github/directory-file-count/voluminor/lightweigit-loader?color=orange)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/voluminor/lightweigit-loader?color=green)
![GitHub repo size](https://img.shields.io/github/repo-size/voluminor/lightweigit-loader)

# LightweiGit Loader

A lightweight, dependency-free Go library for working with **repository tags** and **releases** across different Git hosting platforms.

The library accepts a repository URL, **auto-detects the provider** (GitHub, GitLab, Bitbucket, Gogs-family, etc.), and gives you a unified interface to:

- Find the latest tag / release
- Find a tag / release by name
- Stream tags / releases
- Build source archive URLs (ZIP / TAR) for tags and releases
- Inspect release assets (when available)

## Support status and guarantees

- GitHub support is **100% working and verified**.
- Everything else is included for cross-provider support and future expansion, but is **not verified in real projects yet**.
  - Non-GitHub providers are currently validated only by tests.

## Authentication

This module does not implement authentication.

- No tokens
- No OAuth
- No cookies
- No private repository access helpers

It simply accepts a repository URL and works with public endpoints. If the platform requires authorization for a request, the request will fail and the error will be returned.

## Installation

If you use this repository as a Go module:

```bash
go get github.com/voluminor/lightweigit-loader
````

Note about imports:

* This repository is structured as Go packages (root `lightweigit` plus subpackages).
* Depending on the module path in `go.mod`, imports may look like either:

  * `github.com/voluminor/lightweigit-loader/...`
  * or shorter internal module imports (as used by the repository itself)

In the examples below, replace import paths to match your `go.mod` module path.

## Quick start: detect provider from URL

The simplest entry point is `global.Parse(rawURL)`, which tries known providers in order and returns a `ProviderInterface`.

```go
package main

import (
	"fmt"
	"log"

	"github.com/voluminor/lightweigit-loader/global"
)

func main() {
	raw := "https://github.com/AI-translate-book/template-EN-to-RU/commits/v0.1.0"

	obj, err := global.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Type:", obj.Type())
	fmt.Println("Repo:", obj.String())
	fmt.Println("Domain:", obj.Domain())
	fmt.Println("URL:", obj.URL().String())
}
```

Provider detection is best-effort and supports various URL shapes, including file and commit URLs.

## Core interfaces

The API is intentionally small. Everything revolves around `ProviderInterface`:

* `TagLatest()`

* `TagFind(name string)`

* `TagsStream(ctx, ch, limit)`

* `ReleaseLatest()`

* `ReleaseFind(name string)`

* `ReleasesStream(ctx, ch, limit)`

And the objects returned:

### Tag object

A tag implements:

* `String() string`
* `URL() *url.URL`
* `ZIP() *url.URL`
* `TAR() *url.URL`

### Release object

A release implements:

* `Name() string`
* `BodyMD() string`
* `URL() *url.URL`
* `Tag() ProviderTagInterface`
* `ZIP() *url.URL`
* `TAR() *url.URL`
* `Assets() []ProviderReleaseAssetInterface`
* `IsPrerelease() bool`

### Release asset object

A release asset implements:

* `Name() string`
* `URL() *url.URL`
* `ContentType() string`
* `Size() uint32`

## Working with tags

### Latest tag

```go
package main

import (
	"fmt"
	"log"

	"github.com/voluminor/lightweigit-loader/global"
)

func main() {
	obj, err := global.Parse("https://github.com/OWNER/REPO")
	if err != nil {
		log.Fatal(err)
	}

	tag, err := obj.TagLatest()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Latest tag:", tag.String())
	fmt.Println("Tag URL:", tag.URL().String())
	fmt.Println("ZIP:", tag.ZIP().String())
	fmt.Println("TAR:", tag.TAR().String())
}
```

### Find tag by name

```go
tag, err := obj.TagFind("v1.2.3")
if err != nil {
	// handle not found / provider error
}
fmt.Println(tag.String())
```

### Stream tags

`TagsStream` writes tags into a channel you provide.

A safe consumption pattern is to run streaming in a goroutine and close the channel after it returns:

```go
package main

import (
	"context"
	"fmt"
	"log"

	lightweigit "github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/global"
)

func main() {
	obj, err := global.Parse("https://github.com/OWNER/REPO")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ch := make(chan lightweigit.ProviderTagInterface, 32)

	go func() {
		defer close(ch)
		_ = obj.TagsStream(ctx, ch, 0) // limit=0 means provider-defined/default behavior
	}()

	for tag := range ch {
		fmt.Println(tag.String())
	}
}
```

## Working with releases

### Latest release

```go
rel, err := obj.ReleaseLatest()
if err != nil {
	// handle errors
}

fmt.Println("Release:", rel.Name())
fmt.Println("Tag:", rel.Tag().String())
fmt.Println("Prerelease:", rel.IsPrerelease())
fmt.Println("Release URL:", rel.URL().String())

// Markdown body
fmt.Println(rel.BodyMD())

// Source archives
fmt.Println("ZIP:", rel.ZIP().String())
fmt.Println("TAR:", rel.TAR().String())
```

### Find release by name

```go
rel, err := obj.ReleaseFind("v1.2.3")
if err != nil {
	// handle not found / provider error
}
fmt.Println(rel.Name())
```

### Stream releases

```go
package main

import (
	"context"
	"fmt"
	"log"

	lightweigit "github.com/voluminor/lightweigit-loader"
	"github.com/voluminor/lightweigit-loader/global"
)

func main() {
	obj, err := global.Parse("https://github.com/OWNER/REPO")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ch := make(chan lightweigit.ProviderReleaseInterface, 16)

	go func() {
		defer close(ch)
		_ = obj.ReleasesStream(ctx, ch, 0)
	}()

	for rel := range ch {
		fmt.Println(rel.Name(), rel.Tag().String())
	}
}
```

## Release assets

If the provider exposes release assets, you can inspect them via `Assets()`:

```go
rel, err := obj.ReleaseLatest()
if err != nil {
	log.Fatal(err)
}

for _, a := range rel.Assets() {
	fmt.Println("Asset:", a.Name())
	fmt.Println("  URL:", a.URL().String())
	fmt.Println("  Type:", a.ContentType())
	fmt.Println("  Size:", a.Size())
}
```

## Downloading source archives

Tags and releases both provide archive URLs.

Example download (standard library only):

```go
package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/voluminor/lightweigit-loader/global"
)

func main() {
	obj, err := global.Parse("https://github.com/OWNER/REPO")
	if err != nil {
		log.Fatal(err)
	}

	tag, err := obj.TagFind("v1.2.3")
	if err != nil {
		log.Fatal(err)
	}

	u := tag.ZIP().String()
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	f, err := os.Create("src.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Errors and HTTP behavior

The library provides a shared HTTP client with a short timeout:

* `lightweigit.HttpClient` defaults to 4 seconds
* `lightweigit.ErrNotFound` is returned when the provider responds with HTTP 404

If you need different networking settings (proxy, timeout, transport tuning), you can override the client:

```go
package main

import (
	"net/http"
	"time"

	lightweigit "github.com/voluminor/lightweigit-loader"
)

func main() {
	lightweigit.HttpClient = &http.Client{
		Timeout: 15 * time.Second,
	}
}
```

## Design notes

* Dependency-free (standard library only)
* Provider-first architecture:

  * URL parsing + normalization
  * Provider detection (global parse helper)
  * Provider-specific API implementations behind interfaces
* Built for version-driven workflows:

  * tags/releases as version sources
  * archives as installable source blobs
