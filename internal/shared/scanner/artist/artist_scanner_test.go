package artist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
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

	artists := ScanArtists(conf.MediascanConf{MediaDirs: []string{root}})

	if len(artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(artists))
	}
	a := artists[0]
	if a.Path != artistDir {
		t.Errorf("got path %q, want %q", a.Path, artistDir)
	}
	data := a.ArtistYamlFile.ArtistData
	if len(data.ArtistNames) != 1 || data.ArtistNames[0] != "The Beatles" {
		t.Errorf("unexpected artist names: %#v", data.ArtistNames)
	}
	if data.City != "Liverpool" {
		t.Errorf("got city %q, want %q", data.City, "Liverpool")
	}
	if data.CountryCode != "GB" || data.RegionCode != "ENG" {
		t.Errorf("unexpected codes: %q %q", data.CountryCode, data.RegionCode)
	}
	if len(data.LanguageCodes) != 1 || data.LanguageCodes[0] != "en" {
		t.Errorf("unexpected language codes: %#v", data.LanguageCodes)
	}
	if data.DOB.Y != 1960 || data.DOB.M != 10 || data.DOB.D != 9 {
		t.Errorf("unexpected dob: %#v", data.DOB)
	}
}

func TestScanArtists_SkipsExcludedPaths(t *testing.T) {
	root := t.TempDir()
	writeArtistYml(t, filepath.Join(root, "Keep"), testArtistYml)
	writeArtistYml(t, filepath.Join(root, "ExcludeThisArtistDir"), testArtistYml)

	artists := ScanArtists(conf.MediascanConf{
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

	artists := ScanArtists(conf.MediascanConf{MediaDirs: []string{missing}})

	if len(artists) != 0 {
		t.Fatalf("expected 0 artists, got %d", len(artists))
	}
}
