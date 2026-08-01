package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestApplySEO_UseCasesPage(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	pkgDir := filepath.Dir(thisFile)               // .../cmd/rexec
	repoRoot := filepath.Dir(filepath.Dir(pkgDir)) // .../
	indexPath := filepath.Join(repoRoot, "web", "index.html")

	base, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}

	seo := seoConfig{
		Title:       "Rexec Use Cases - The Future of Development",
		Description: "Discover how Rexec powers ephemeral development environments, AI agent execution, collaborative coding, and secure cloud access.",
	}
	canonical := "https://rexec.pipeops.io/use-cases"

	if !reMetaContent("name", "description").Match(base) {
		t.Fatalf("base index.html missing meta name=description")
	}
	if !reMetaContent("property", "og:description").Match(base) {
		t.Fatalf("base index.html missing meta property=og:description")
	}

	out := applySEO(string(base), seo, canonical)

	if !strings.Contains(out, `name="description"`) {
		t.Fatalf(`missing meta name="description" after applySEO`)
	}

	if !strings.Contains(out, "<title>Rexec Use Cases - The Future of Development</title>") {
		t.Fatalf("missing or incorrect <title> tag in SEO output")
	}

	if !strings.Contains(out, "Discover how Rexec powers ephemeral development environments") {
		t.Fatalf("expected description not present in output")
	}

	descMeta := reMetaContent("name", "description").FindStringSubmatch(out)
	if len(descMeta) != 4 {
		t.Fatalf("failed to extract meta description content")
	}
	if descMeta[2] != seo.Description {
		t.Fatalf("meta description mismatch: got %q", descMeta[2])
	}

	mustMatch := func(pattern string) {
		re := regexp.MustCompile(pattern)
		if !re.MatchString(out) {
			t.Fatalf("missing expected pattern: %s", pattern)
		}
	}

	mustMatch(`(?is)<meta[^>]+name=["']description["'][^>]+content=["']Discover how Rexec powers ephemeral development environments, AI agent execution, collaborative coding, and secure cloud access\.[\"']`)
	mustMatch(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']Rexec Use Cases - The Future of Development[\"']`)
	mustMatch(`(?is)<meta[^>]+property=["']og:description["'][^>]+content=["']Discover how Rexec powers ephemeral development environments, AI agent execution, collaborative coding, and secure cloud access\.[\"']`)
	mustMatch(`(?is)<meta[^>]+property=["']og:url["'][^>]+content=["']https://rexec\.pipeops\.io/use-cases[\"']`)
	mustMatch(`(?is)<meta[^>]+name=["']twitter:title["'][^>]+content=["']Rexec Use Cases - The Future of Development[\"']`)
	mustMatch(`(?is)<meta[^>]+name=["']twitter:description["'][^>]+content=["']Discover how Rexec powers ephemeral development environments, AI agent execution, collaborative coding, and secure cloud access\.[\"']`)
	mustMatch(`(?is)<meta[^>]+name=["']twitter:url["'][^>]+content=["']https://rexec\.pipeops\.io/use-cases[\"']`)
	mustMatch(`(?is)<link[^>]+rel=["']canonical["'][^>]+href=["']https://rexec\.pipeops\.io/use-cases[\"']`)
}

func TestApplySEO_CustomFields(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	pkgDir := filepath.Dir(thisFile)               // .../cmd/rexec
	repoRoot := filepath.Dir(filepath.Dir(pkgDir)) // .../
	indexPath := filepath.Join(repoRoot, "web", "index.html")

	base, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}

	seo := seoConfig{
		Title:              "Page Title",
		Description:        "Page Description",
		OGTitle:            "OG Title",
		OGDescription:      "OG Description",
		OGType:             "article",
		TwitterTitle:       "Twitter Title",
		TwitterDescription: "Twitter Description",
	}
	canonical := "https://rexec.pipeops.io/use-cases/ephemeral-dev-environments"

	out := applySEO(string(base), seo, canonical)

	if !strings.Contains(out, "<title>Page Title</title>") {
		t.Fatalf("missing or incorrect <title> tag in SEO output")
	}

	mustEqual := func(re *regexp.Regexp, expected string) {
		m := re.FindStringSubmatch(out)
		if len(m) != 4 {
			t.Fatalf("expected match for %v", re)
		}
		if m[2] != expected {
			t.Fatalf("expected %q, got %q", expected, m[2])
		}
	}

	mustEqual(reMetaContent("name", "description"), "Page Description")
	mustEqual(reMetaContent("property", "og:title"), "OG Title")
	mustEqual(reMetaContent("property", "og:description"), "OG Description")
	mustEqual(reMetaContent("property", "og:type"), "article")
	mustEqual(reMetaContent("property", "og:url"), canonical)
	mustEqual(reMetaContent("name", "twitter:title"), "Twitter Title")
	mustEqual(reMetaContent("name", "twitter:description"), "Twitter Description")
	mustEqual(reMetaContent("name", "twitter:url"), canonical)

	m := reCanonicalHref.FindStringSubmatch(out)
	if len(m) != 4 {
		t.Fatalf("expected match for canonical link")
	}
	if m[2] != canonical {
		t.Fatalf("expected canonical %q, got %q", canonical, m[2])
	}
}
