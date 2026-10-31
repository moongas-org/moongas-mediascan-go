package album

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
)

const testAlbumYml = `albumData:
  name: Example Album
  dateReleased:
    y: 1993
    m: 5
    d: 4
  explicitCover: true
`

func writeAlbumYml(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "album.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanAlbumsReadsOnlyAlbumYml(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	dir := filepath.Join(first, "Artist", "Album")
	writeAlbumYml(t, dir, testAlbumYml)
	writeAlbumYml(t, filepath.Join(second, "Other"), "albumData:\n  name: Other Album\n  dateReleased:\n    y: 2000\n")
	for _, name := range []string{"artist.yml", "album.yaml", "notes.txt", "song.mp3"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("invalid: ["), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "album.yml.directory"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := ScanAlbums(conf.MediascanConf{MediaDirs: []string{first, second}})
	want := []album.Album{
		{
			AlbumYamlFile: album.AlbumYamlFile{
				AlbumData: album.AlbumData{Name: "Example Album", DateReleased: date.Date{Y: 1993, M: 5, D: 4}, ExplicitCover: true},
			},
			Path: dir,
		},
		{
			AlbumYamlFile: album.AlbumYamlFile{
				AlbumData: album.AlbumData{Name: "Other Album", DateReleased: date.Date{Y: 2000}},
			},
			Path: filepath.Join(second, "Other"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("albums = %+v, want %+v", got, want)
	}
}

func TestScanAlbumsSkipsExcludedPaths(t *testing.T) {
	root := t.TempDir()
	writeAlbumYml(t, filepath.Join(root, "Keep"), testAlbumYml)
	writeAlbumYml(t, filepath.Join(root, "OmitThisAlbum", "Nested"), testAlbumYml)
	got := ScanAlbums(conf.MediascanConf{
		MediaDirs: []string{root}, ExcludePaths: []string{"OMITTHISALBUM"},
	})
	if len(got) != 1 || got[0].AlbumYamlFile.AlbumData.Name != "Example Album" ||
		got[0].Path != filepath.Join(root, "Keep") {
		t.Fatalf("expected one included album, got %+v", got)
	}
}

func TestScanAlbumsLogsErrorsAndContinues(t *testing.T) {
	root := t.TempDir()
	writeAlbumYml(t, filepath.Join(root, "Invalid"), "albumData: [")
	writeAlbumYml(t, filepath.Join(root, "Valid"), testAlbumYml)
	missing := filepath.Join(root, "does-not-exist")
	broken := filepath.Join(root, "Broken")
	if err := os.Mkdir(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(missing, filepath.Join(broken, "album.yml")); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	original := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(original) })

	got := ScanAlbums(conf.MediascanConf{MediaDirs: []string{missing, root}})
	if len(got) != 1 || got[0].AlbumYamlFile.AlbumData.Name != "Example Album" ||
		got[0].Path != filepath.Join(root, "Valid") {
		t.Fatalf("expected valid album despite errors, got %+v", got)
	}
	for _, message := range []string{"ERROR accessing " + missing, "ERROR parsing ", "ERROR reading "} {
		if !strings.Contains(output.String(), message) {
			t.Errorf("missing log %q in %s", message, output.String())
		}
	}
}

func TestScanAlbumsEmptyLibrary(t *testing.T) {
	for _, dirs := range [][]string{nil, {t.TempDir()}} {
		if got := ScanAlbums(conf.MediascanConf{MediaDirs: dirs}); len(got) != 0 {
			t.Fatalf("expected no albums, got %+v", got)
		}
	}
}
