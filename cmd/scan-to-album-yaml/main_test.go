package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/cli"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
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
	if err := writeAlbumYamlFiles(files, cli.SkipExisting); err != nil {
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
			if err := writeAlbumYamlFiles(files, cli.SkipExisting); err != nil {
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
	if err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}}, cli.SkipExisting); err != nil {
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
	if err := writeAlbumYamlFiles(files, cli.OverwriteExisting); err != nil {
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
	if err := writeAlbumYamlFiles(files, cli.OverwriteExisting); err != nil {
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
	if err := writeAlbumYamlFiles(files, cli.OverwriteExisting); err != nil {
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
	if err := writeAlbumYamlFiles(nil, cli.OverwriteNewer); err == nil {
		t.Fatal("expected unsupported overwrite mode error")
	}
}

func TestWriteAlbumYamlFilesEmpty(t *testing.T) {
	if err := writeAlbumYamlFiles(nil, cli.SkipExisting); err != nil {
		t.Fatal(err)
	}
}

func TestWriteAlbumYamlFilesReportsCreateError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}}, cli.SkipExisting)
	if err == nil || !strings.Contains(err.Error(), filepath.Join(dir, "album.yml")) {
		t.Fatalf("expected create error with album.yml path, got %v", err)
	}
}

func TestWriteAlbumYamlFilesRejectsMissingPath(t *testing.T) {
	for _, path := range []*string{nil, utils.Pointer("")} {
		err := writeAlbumYamlFiles([]media_file.MediaFileData{{Path: path}}, cli.SkipExisting)
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
	if err := writeAlbumYamlFiles(files, cli.SkipExisting); err != nil {
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
