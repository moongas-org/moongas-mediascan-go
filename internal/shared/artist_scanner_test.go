package shared

import (
	"os"
	"path/filepath"
	"testing"
)

const testArtistYml = `artistData:
  artistNames:
    - The Beatles
  city: Liverpool
  countryCode: GB
  regionCode: ENG
  languageCodes:
    - en
  dob:
    y: 1960
    m: 10
    d: 9
`

func writeArtistYml(t *testing.T, dir string, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create artist dir: %v", err)
	}
	path := filepath.Join(dir, "artist.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artist.yml: %v", err)
	}
	return path
}

func TestScanArtists_ReadsArtistYmlFiles(t *testing.T) {
	root := t.TempDir()
	artistDir := filepath.Join(root, "Music", "Beatles")
	writeArtistYml(t, artistDir, testArtistYml)
	// A non-artist.yml file must be ignored.
	if err := os.WriteFile(filepath.Join(artistDir, "notes.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatalf("write notes: %v", err)
	}

	artists := ScanArtists(MediascanConf{MediaDirs: []string{root}})

	if len(artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(artists))
	}
	a := artists[0]
	if a.Path != artistDir {
		t.Errorf("got path %q, want %q", a.Path, artistDir)
	}
	if len(a.ArtistData.ArtistNames) != 1 || a.ArtistData.ArtistNames[0] != "The Beatles" {
		t.Errorf("unexpected artist names: %#v", a.ArtistData.ArtistNames)
	}
	if a.ArtistData.City != "Liverpool" {
		t.Errorf("got city %q, want %q", a.ArtistData.City, "Liverpool")
	}
	if a.ArtistData.CountryCode != "GB" || a.ArtistData.RegionCode != "ENG" {
		t.Errorf("unexpected codes: %q %q", a.ArtistData.CountryCode, a.ArtistData.RegionCode)
	}
	if len(a.ArtistData.LanguageCodes) != 1 || a.ArtistData.LanguageCodes[0] != "en" {
		t.Errorf("unexpected language codes: %#v", a.ArtistData.LanguageCodes)
	}
	if a.ArtistData.DOB.Y != 1960 || a.ArtistData.DOB.M != 10 || a.ArtistData.DOB.D != 9 {
		t.Errorf("unexpected dob: %#v", a.ArtistData.DOB)
	}
}

func TestScanArtists_SkipsExcludedPaths(t *testing.T) {
	root := t.TempDir()
	writeArtistYml(t, filepath.Join(root, "Keep"), testArtistYml)
	writeArtistYml(t, filepath.Join(root, "ExcludeThisArtistDir"), testArtistYml)

	artists := ScanArtists(MediascanConf{
		MediaDirs:    []string{root},
		ExcludePaths: []string{"ExcludeThisArtistDir"},
	})

	if len(artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(artists))
	}
	if filepath.Base(artists[0].Path) != "Keep" {
		t.Errorf("expected kept artist dir, got %q", artists[0].Path)
	}
}

func TestScanArtists_SkipsMissingMediaDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	artists := ScanArtists(MediascanConf{MediaDirs: []string{missing}})

	if len(artists) != 0 {
		t.Fatalf("expected 0 artists, got %d", len(artists))
	}
}
