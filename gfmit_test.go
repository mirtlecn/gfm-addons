package gfmit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"
)

type packageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func expectedRemoteURL(t *testing.T, file string) string {
	t.Helper()

	data, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("ReadFile(package.json) error = %v", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatalf("Unmarshal(package.json) error = %v", err)
	}

	return fmt.Sprintf("https://cdn.jsdelivr.net/npm/%s@%s/%s", pkg.Name, pkg.Version, minifiedAssetPath(file))
}

func minifiedAssetPath(file string) string {
	extension := path.Ext(file)
	if extension == "" {
		return file
	}
	return file[:len(file)-len(extension)] + ".min" + extension
}

func TestAssets(t *testing.T) {
	items := Assets()
	if got, want := len(items), 10; got != want {
		t.Fatalf("len(Assets()) = %d, want %d", got, want)
	}

	first, ok := GetAsset("ravel_gfm_css")
	if !ok {
		t.Fatal("ravel_gfm_css is missing")
	}
	if first.File != "assets/ravel-gfm.css" {
		t.Fatalf("ravel_gfm_css file = %q", first.File)
	}
	if first.ContentType != "text/css; charset=utf-8" {
		t.Fatalf("ravel_gfm_css content type = %q", first.ContentType)
	}
	if first.RemoteURL != expectedRemoteURL(t, "assets/ravel-gfm.css") {
		t.Fatalf("ravel_gfm_css remote URL = %q", first.RemoteURL)
	}

	if _, ok := GetAsset("missing"); ok {
		t.Fatal("missing asset unexpectedly exists")
	}

	highlight, ok := GetAsset("highlight_js")
	if !ok {
		t.Fatal("highlight_js is missing")
	}
	if highlight.ContentType != "application/javascript; charset=utf-8" {
		t.Fatalf("highlight_js content type = %q", highlight.ContentType)
	}

	folio, ok := GetAsset("folio_gfm_css")
	if !ok {
		t.Fatal("folio_gfm_css is missing")
	}
	if folio.File != "assets/folio-gfm.css" {
		t.Fatalf("folio_gfm_css file = %q", folio.File)
	}
	if folio.RemoteURL != expectedRemoteURL(t, "assets/folio-gfm.css") {
		t.Fatalf("folio_gfm_css remote URL = %q", folio.RemoteURL)
	}
}

func TestReadAsset(t *testing.T) {
	content, asset, err := ReadAsset("github_gfm_css")
	if err != nil {
		t.Fatalf("ReadAsset() error = %v", err)
	}
	if asset.File != "assets/github-gfm.css" {
		t.Fatalf("github_gfm_css file = %q", asset.File)
	}
	if len(content) == 0 {
		t.Fatal("github_gfm_css content is empty")
	}
}

func TestEmbeddedAssetContentIsGeneratedSeparatelyFromPackagedSource(t *testing.T) {
	sourceContent, _, err := ReadAsset("github_gfm_css")
	if err != nil {
		t.Fatalf("ReadAsset() error = %v", err)
	}
	embeddedContent, err := ReadEmbeddedAssetContent("github_gfm_css")
	if err != nil {
		t.Fatalf("ReadEmbeddedAssetContent() error = %v", err)
	}
	if len(embeddedContent) == 0 {
		t.Fatal("embedded asset content is empty")
	}
	if len(embeddedContent) > len(sourceContent) {
		t.Fatalf("embedded asset length = %d, source length = %d", len(embeddedContent), len(sourceContent))
	}

	html, err := RenderMarkdownToHTML("# Hello", RenderOptions{AssetMode: "inline", CSS: "github_gfm_css"})
	if err != nil {
		t.Fatalf("RenderMarkdownToHTML() error = %v", err)
	}
	if !bytes.Contains([]byte(html), embeddedContent) {
		t.Fatal("inline HTML does not contain generated embedded asset content")
	}
}

func TestReadEmbeddedAssetContentRejectsUnknownAsset(t *testing.T) {
	if _, err := ReadEmbeddedAssetContent("missing"); err == nil {
		t.Fatal("ReadEmbeddedAssetContent() unexpectedly succeeded")
	}
}

func TestAssetsReturnsCopy(t *testing.T) {
	items := Assets()
	items[0].Key = "changed"
	if MustAsset("ravel_gfm_css").Key != "ravel_gfm_css" {
		t.Fatal("Assets returned mutable package state")
	}
}
