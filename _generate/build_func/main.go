package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dep "github.com/voluminor/lightweigit-loader/_generate"
)

// // // // // // // // // //

const (
	packageName = "global"
	fileName    = "func.go"

	fTag  = "func_tag.go"
	ptTag = "TagLatest() (lightweigit.ProviderTagInterface, error)"

	fRelease  = "func_release.go"
	ptRelease = "ReleaseLatest() (lightweigit.ProviderReleaseInterface, error)"
)

//go:embed template.tmpl
var template_text string

//

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func fileContains(path, substr string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(b), substr), nil
}

func FindValidDirs(rootDir string) ([]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("dir %q: %w", rootDir, err)
	}

	var result []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		dirPath := filepath.Join(rootDir, dirName)

		tagFile := filepath.Join(dirPath, fTag)
		releaseFile := filepath.Join(dirPath, fRelease)

		if !fileExists(tagFile) || !fileExists(releaseFile) {
			continue
		}

		okTag, err := fileContains(tagFile, ptTag)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", tagFile, err)
		}
		if !okTag {
			continue
		}

		okRelease, err := fileContains(releaseFile, ptRelease)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", releaseFile, err)
		}
		if !okRelease {
			continue
		}

		result = append(result, dirName)
	}

	return result, nil
}

type TemplateObj struct {
	GenerationTime string
	PackageName    string
	Path           string
	ImportsArr     []string

	Mods []string
	Dirs []string
}

// //

func main() {

	dirs, err := FindValidDirs(".")
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(dirs)

	// //

	data := new(TemplateObj)
	data.GenerationTime = time.Now().Format(time.RFC3339)
	data.PackageName = packageName

	data.ImportsArr = make([]string, 0)
	data.ImportsArr = append(data.ImportsArr, "errors")
	data.ImportsArr = append(data.ImportsArr, "github.com/voluminor/lightweigit-loader")

	data.Dirs = dirs
	for _, dir := range dirs {
		data.ImportsArr = append(data.ImportsArr, "github.com/voluminor/lightweigit-loader/"+dir)
	}

	os.MkdirAll(filepath.Join("target", packageName), 0777)
	err = dep.WriteFileFromTemplate(filepath.Join("target", packageName, fileName), template_text, data)
	if err != nil {
		log.Println("An error when trying to save a generated file:", err)
	}
}
