package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared"
	"gopkg.in/yaml.v3"
)

func TestWriteAlbumYamlFiles(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	files := []shared.MediaFileData{
		{AlbumPath: first, Album: "First album", Year: 1993},
		{AlbumPath: second, Album: "Other album", Year: 2000},
		{AlbumPath: first, Album: "Later tags", Year: 2020},
	}
	if err := writeAlbumYamlFiles(files); err != nil {
		t.Fatal(err)
	}
	for _, m := range files[:2] {
		path := filepath.Join(m.AlbumPath, "album.yml")
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
				"name":          m.Album,
				"dateReleased":  map[string]interface{}{"y": m.Year},
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
			files := []shared.MediaFileData{{AlbumPath: dir, Album: "Replacement", Year: 2020}}
			if err := writeAlbumYamlFiles(files); err != nil {
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
	if err := writeAlbumYamlFiles([]shared.MediaFileData{{AlbumPath: dir}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("symlink target should remain missing, got %v", err)
	}
}

func TestWriteAlbumYamlFilesEmpty(t *testing.T) {
	if err := writeAlbumYamlFiles(nil); err != nil {
		t.Fatal(err)
	}
}

func TestWriteAlbumYamlFilesReportsCreateError(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	err := writeAlbumYamlFiles([]shared.MediaFileData{{AlbumPath: dir}})
	if err == nil || !strings.Contains(err.Error(), filepath.Join(dir, "album.yml")) {
		t.Fatalf("expected create error with album.yml path, got %v", err)
	}
}
