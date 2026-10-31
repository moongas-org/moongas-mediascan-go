package track

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
)

const trackYAML = `trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: Album
  genre: Rock
  year: 2026
`

func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanTracksAlbumLevelOnly(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	wantPaths := []string{
		filepath.Join(first, "Artist", "Album", "song.yml"),
		filepath.Join(second, "Artist", "Album", "other.yml"),
	}
	for _, path := range wantPaths {
		writeYAML(t, path, trackYAML)
	}
	for _, relative := range []string{
		"root.yml", "Artist/artist-track.yml", "Artist/Album/Nested/song.yml",
		"Artist/Album/song.yaml", "Artist/Album/artist.yml", "Artist/Album/album.yml",
		"Artist/Album/song.mp3",
	} {
		writeYAML(t, filepath.Join(first, relative), trackYAML)
	}
	got := ScanTracks(conf.MediascanConf{MediaDirs: []string{first, second}})
	if len(got) != len(wantPaths) {
		t.Fatalf("expected %d tracks, got %+v", len(wantPaths), got)
	}
	for i, want := range wantPaths {
		if got[i].Path != want || got[i].TrackYamlFile.TrackData.Title != "Song" ||
			got[i].TrackYamlFile.TrackData.Year != 2026 {
			t.Errorf("unexpected track: %+v", got[i])
		}
	}
}

func TestScanTracksExcludedPaths(t *testing.T) {
	root := t.TempDir()
	writeYAML(t, filepath.Join(root, "Artist", "Keep", "song.yml"), trackYAML)
	writeYAML(t, filepath.Join(root, "Artist", "OmitThisAlbum", "song.yml"), trackYAML)
	got := ScanTracks(conf.MediascanConf{
		MediaDirs: []string{root}, ExcludePaths: []string{"OMITTHISALBUM"},
	})
	if len(got) != 1 || filepath.Base(filepath.Dir(got[0].Path)) != "Keep" {
		t.Fatalf("unexpected tracks: %+v", got)
	}
}

func TestScanTracksLogsErrorsAndContinues(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	writeYAML(t, filepath.Join(dir, "valid.yml"), trackYAML)
	writeYAML(t, filepath.Join(dir, "invalid.yml"), "trackData: [")
	writeYAML(t, filepath.Join(dir, "missing-fields.yml"), "trackData:\n  title: Song\n")
	missing := filepath.Join(root, "missing")
	if err := os.Symlink(missing, filepath.Join(dir, "unreadable.yml")); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	original := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(original) })
	got := ScanTracks(conf.MediascanConf{MediaDirs: []string{missing, root}})
	if len(got) != 1 || filepath.Base(got[0].Path) != "valid.yml" {
		t.Fatalf("expected valid track despite failures: %+v", got)
	}
	for _, message := range []string{"ERROR accessing ", "ERROR reading ", "ERROR parsing ", "missing required field"} {
		if !strings.Contains(output.String(), message) {
			t.Errorf("missing log %q in %s", message, output.String())
		}
	}
}

func TestScanTracksEmptyLibrary(t *testing.T) {
	for _, dirs := range [][]string{nil, {t.TempDir()}} {
		if got := ScanTracks(conf.MediascanConf{MediaDirs: dirs}); len(got) != 0 {
			t.Fatalf("expected no tracks, got %+v", got)
		}
	}
}
