package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
)

func newAlbumTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestInsertAlbums(t *testing.T) {
	for _, tt := range []struct {
		name         string
		yaml         string
		wantName     string
		wantYear     int
		wantExplicit bool
	}{
		{
			name:     "missing YAML uses first media file",
			wantName: "Tagged album",
			wantYear: 1980,
		},
		{
			name: "YAML overrides media tags",
			yaml: `albumData:
  name: Matando güreos
  dateReleased:
    y: 1993
    m: 5
    d: 4
  explicitCover: true
`,
			wantName:     "Matando güreos",
			wantYear:     1993,
			wantExplicit: true,
		},
		{
			name: "YAML cover defaults to false",
			yaml: `albumData:
  name: YAML album
  dateReleased:
    y: 2000
`,
			wantName: "YAML album",
			wantYear: 2000,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := newAlbumTestDB(t)
			dir := t.TempDir()
			if tt.yaml != "" {
				if err := os.WriteFile(filepath.Join(dir, "album.yml"), []byte(tt.yaml), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			files := []media_file.MediaFileData{
				{AlbumPath: dir, Album: "Tagged album", Year: 1980},
				{AlbumPath: dir, Album: "Later album", Year: 2020},
			}
			if err := insertAlbums(db, files); err != nil {
				t.Fatal(err)
			}

			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("got %d albums, want 1", count)
			}
			var id, year int
			var path, name string
			var explicit bool
			if err := db.QueryRow("SELECT id, path, name, year, explicitcover FROM album").
				Scan(&id, &path, &name, &year, &explicit); err != nil {
				t.Fatal(err)
			}
			if id != 1 || path != dir || name != tt.wantName || year != tt.wantYear || explicit != tt.wantExplicit {
				t.Errorf("unexpected album: id=%d path=%q name=%q year=%d explicitcover=%t",
					id, path, name, year, explicit)
			}
		})
	}
}

func TestInsertAlbumsUniquePaths(t *testing.T) {
	db := newAlbumTestDB(t)
	first, second := t.TempDir(), t.TempDir()
	files := []media_file.MediaFileData{
		{AlbumPath: first, Album: "Same name", Year: 1990},
		{AlbumPath: second, Album: "Same name", Year: 1990},
		{AlbumPath: first, Album: "Duplicate", Year: 2000},
	}
	if err := insertAlbums(db, files); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("got %d albums, want 2", count)
	}
	if _, err := db.Exec("INSERT INTO album(path) VALUES (?)", first); err == nil {
		t.Error("expected duplicate album path to be rejected")
	}
}

func TestInsertAlbumsEmpty(t *testing.T) {
	db := newAlbumTestDB(t)
	if err := insertAlbums(db, nil); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("got %d albums, want 0", count)
	}
}

func TestInsertAlbumsReportsMetadataErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		dir  bool
		want string
	}{
		{name: "invalid YAML", want: "parse "},
		{name: "unreadable YAML path", dir: true, want: "read "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := newAlbumTestDB(t)
			dir := t.TempDir()
			path := filepath.Join(dir, "album.yml")
			if tt.dir {
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte("albumData: ["), 0o644); err != nil {
				t.Fatal(err)
			}
			err := insertAlbums(db, []media_file.MediaFileData{{AlbumPath: dir}})
			if err == nil || !strings.Contains(err.Error(), tt.want+path) {
				t.Fatalf("got error %v, want %q with metadata path", err, tt.want)
			}
		})
	}
}
