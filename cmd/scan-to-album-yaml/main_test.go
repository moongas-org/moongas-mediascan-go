package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

func TestWriteAlbumYamlFiles(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	files := []media_file.MediaFileData{
		{Path: utils.Pointer(filepath.Join(first, "first.mp3")), Album: utils.Pointer("First album"), Year: utils.Pointer(1993)},
		{Path: utils.Pointer(filepath.Join(second, "other.mp3")), Album: utils.Pointer("Other album"), Year: utils.Pointer(2000)},
		{Path: utils.Pointer(filepath.Join(first, "later.mp3")), Album: utils.Pointer("Later tags"), Year: utils.Pointer(2020)},
	}
	if err := writeAlbumYamlFiles(files, nil, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
	for _, m := range files[:2] {
		path := filepath.Join(filepath.Dir(*m.Path), "album.yml")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]interface{}
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]interface{}{
			"albumData": map[string]interface{}{
				"name":          *m.Album,
				"dateReleased":  map[string]interface{}{"y": *m.Year},
				"explicitCover": false,
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("unexpected YAML in %s: got %#v, want %#v", path, got, want)
		}
	}
}

func TestWriteAlbumYamlFilesPreservesExisting(t *testing.T) {
	for _, content := range []string{
		"albumData:\n  name: Custom\n  explicitCover: true\n",
		"not valid YAML: [",
		"",
	} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "album.yml")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			files := []media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3")), Album: utils.Pointer("Replacement"), Year: utils.Pointer(2020)}}
			if err := writeAlbumYamlFiles(files, nil, cli.SkipExisting); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			after, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != content || !before.ModTime().Equal(after.ModTime()) {
				t.Error("existing album.yml was modified")
			}
		})
	}
}

func TestWriteAlbumYamlFilesPreservesDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "missing.yml")
	path := filepath.Join(dir, "album.yml")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}}, nil, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("symlink target should remain missing, got %v", err)
	}
}

func TestWriteAlbumYamlFilesOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album.yml")
	if err := os.WriteFile(path, []byte("albumData:\n  name: Old name\n  explicitCover: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := []media_file.MediaFileData{{
		Path:  utils.Pointer(filepath.Join(dir, "song.mp3")),
		Album: utils.Pointer("Replacement"), Year: utils.Pointer(2026), Genre: utils.Pointer("Rock"),
	}}
	if err := writeAlbumYamlFiles(files, nil, cli.OverwriteExisting); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got album.AlbumYamlFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.AlbumData.Name != "Replacement" || got.AlbumData.DateReleased.Y != 2026 ||
		!reflect.DeepEqual(got.AlbumData.Genres, []string{"Rock"}) || !got.AlbumData.ExplicitCover ||
		strings.Contains(string(data), "Old name") {
		t.Fatalf("unexpected replacement YAML: %s", data)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := writeAlbumYamlFiles(files, nil, cli.OverwriteExisting); err != nil {
		t.Fatal(err)
	}
}

func TestWriteAlbumYamlFilesOverwritePreservesFalseExplicitCover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album.yml")
	if err := os.WriteFile(path, []byte("albumData:\n  explicitCover: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := []media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}}
	if err := writeAlbumYamlFiles(files, nil, cli.OverwriteExisting); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got album.AlbumYamlFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.AlbumData.ExplicitCover {
		t.Fatal("expected explicitCover to remain false")
	}
}

func TestWriteAlbumYamlFilesOverwriteRejectsInvalidExistingYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album.yml")
	content := "albumData: ["
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	err := writeAlbumYamlFiles(
		[]media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}},
		nil,
		cli.OverwriteExisting,
	)
	if err == nil || !strings.Contains(err.Error(), "preserve explicitCover") {
		t.Fatalf("expected explicitCover preservation parse error, got %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Fatal("invalid existing YAML should not be overwritten")
	}
}

func TestWriteAlbumYamlFilesRejectsOverwriteNewer(t *testing.T) {
	if err := writeAlbumYamlFiles(nil, nil, cli.OverwriteNewer); err == nil {
		t.Fatal("expected unsupported overwrite mode error")
	}
}

func TestWriteAlbumYamlFilesEmpty(t *testing.T) {
	if err := writeAlbumYamlFiles(nil, nil, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
}

func TestWriteAlbumYamlFilesReportsCreateError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}}, nil, cli.SkipExisting)
	if err == nil || !strings.Contains(err.Error(), filepath.Join(dir, "album.yml")) {
		t.Fatalf("expected create error with album.yml path, got %v", err)
	}
}

func TestWriteAlbumYamlFilesRejectsMissingPath(t *testing.T) {
	for _, path := range []*string{nil, utils.Pointer("")} {
		err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: path}}, nil, cli.SkipExisting)
		if err == nil || !strings.Contains(err.Error(), "missing a path") {
			t.Fatalf("expected missing path error, got %v", err)
		}
	}

}

func TestWriteAlbumYamlFilesCollectsUniqueGenres(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	files := []media_file.MediaFileData{
		{Path: utils.Pointer(filepath.Join(first, "01.mp3")), Genre: utils.Pointer("Rock")},
		{Path: utils.Pointer(filepath.Join(second, "01.mp3")), Genre: utils.Pointer("Jazz")},
		{Path: utils.Pointer(filepath.Join(first, "02.mp3")), Genre: utils.Pointer("Zydeco")},
		{Path: utils.Pointer(filepath.Join(first, "03.mp3")), Genre: utils.Pointer("Rock")},
		{Path: utils.Pointer(filepath.Join(first, "04.mp3")), Genre: utils.Pointer("")},
		{Path: utils.Pointer(filepath.Join(first, "05.mp3"))},
		{Path: utils.Pointer(filepath.Join(first, "06.mp3")), Genre: utils.Pointer("rock")},
		{Path: utils.Pointer(filepath.Join(second, "02.mp3")), Genre: utils.Pointer("Jazz")},
	}
	if err := writeAlbumYamlFiles(files, nil, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
	for dir, want := range map[string][]string{
		first:  {"Rock", "Zydeco", "rock"},
		second: {"Jazz"},
	} {
		data, err := os.ReadFile(filepath.Join(dir, "album.yml"))
		if err != nil {
			t.Fatal(err)
		}
		var got album.AlbumYamlFile
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.AlbumData.Genres, want) {
			t.Errorf("genres for %s = %v, want %v", dir, got.AlbumData.Genres, want)
		}
	}
}

func TestWriteAlbumYamlFilesPrefersTrackYaml(t *testing.T) {
	dir := t.TempDir()
	media := func(name, genre string) media_file.MediaFileData {
		return media_file.MediaFileData{
			Path: utils.Pointer(filepath.Join(dir, name)), Album: utils.Pointer("Media album"),
			Year: utils.Pointer(2000), Genre: utils.Pointer(genre),
		}
	}
	files := []media_file.MediaFileData{
		media("01-media-only.mp3", "Rock"),
		media("02-sidecar.mp3", "Wrong sidecar genre"),
		media("03-renamed.mp3", "Wrong relative path genre"),
		media("04-renamed.flac", "Wrong absolute path genre"),
		media("05-media-only.mp3", "Jazz"),
	}
	tracks := []track.Track{
		{
			Path: filepath.Join(dir, "02-sidecar.yml"),
			TrackYamlFile: track.TrackYamlFile{
				TrackData: track.TrackData{Album: "YAML album", Year: 2024, Genre: "Jazz"},
			},
		},
		{
			Path: filepath.Join(dir, "03-different-name.yml"),
			TrackYamlFile: track.TrackYamlFile{
				TrackData: track.TrackData{Album: "Later YAML album", Year: 2025, Genre: "Folk"},
				MediaFileData: []media_file.MediaFileData{
					{Path: utils.Pointer("./03-renamed.mp3")},
					{Path: utils.Pointer(filepath.Join(dir, "04-renamed.flac"))},
				},
			},
		},
	}
	if err := writeAlbumYamlFiles(files, tracks, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
	assertAlbumYaml(t, dir, "YAML album", 2024, []string{"Jazz", "Folk", "Rock"})
}

func TestScanToAlbumYamlWithoutMediaFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: YAML-only album
  genre: Jazz
  year: 2026
`
	if err := os.WriteFile(filepath.Join(dir, "Song.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	conf := conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}}
	if err := scanToAlbumYaml(conf, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
	assertAlbumYaml(t, dir, "YAML-only album", 2026, []string{"Jazz"})
}

func assertAlbumYaml(t *testing.T, dir, name string, year int, genres []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "album.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var got album.AlbumYamlFile
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.AlbumData.Name != name || got.AlbumData.DateReleased.Y != year ||
		!reflect.DeepEqual(got.AlbumData.Genres, genres) {
		t.Fatalf("got:\n%s\nwant name %q, year %d, genres %v", data, name, year, genres)
	}
}
