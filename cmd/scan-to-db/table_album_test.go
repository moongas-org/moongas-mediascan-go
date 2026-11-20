package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/album"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/date"
)

func newAlbumTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestInsertAlbums(t *testing.T) {
	db := newAlbumTestDB(t)
	artistPath := filepath.Join(t.TempDir(), "Artist")
	if err := insertArtist(db, "/unrelated", "Other", "", "", "", "", true); err != nil {
		t.Fatal(err)
	}
	if err := insertArtist(db, artistPath, "Artist", "", "", "", "", true); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(artistPath, "Album")
	albums := []album.Album{
		{Path: path, HasAlbumYaml: true, AlbumYamlFile: album.AlbumYamlFile{AlbumData: album.AlbumData{
			Name: "YAML album", DateReleased: date.Date{Y: 1993}, ExplicitCover: true,
			Genres: []string{"Zydeco", "Rock"},
		}}},
		{Path: path},
		{Path: filepath.Join(artistPath, "Second")},
	}
	if err := insertAlbums(db, albums); err != nil {
		t.Fatal(err)
	}
	var name, genre string
	var year, artistID, count int
	var explicit, hasAlbumYaml bool
	if err := db.QueryRow("SELECT name, year, explicitcover, artistid, genre, hasalbumyaml FROM album WHERE path = ?", path).
		Scan(&name, &year, &explicit, &artistID, &genre, &hasAlbumYaml); err != nil {
		t.Fatal(err)
	}
	if name != "YAML album" || year != 1993 || !explicit || artistID != 2 || !hasAlbumYaml {
		t.Fatalf("unexpected album: %s %d %t artistid=%d hasalbumyaml=%t", name, year, explicit, artistID, hasAlbumYaml)
	}
	if genre != "Zydeco" {
		t.Fatalf("genre = %q, want first genre Zydeco", genre)
	}
	if err := db.QueryRow("SELECT name, genre, hasalbumyaml FROM album WHERE path = ?", filepath.Join(artistPath, "Second")).
		Scan(&name, &genre, &hasAlbumYaml); err != nil {
		t.Fatal(err)
	}
	if name != "Second" || genre != "" || hasAlbumYaml {
		t.Fatalf("unexpected fallback album: name=%q genre=%q hasalbumyaml=%t", name, genre, hasAlbumYaml)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected two unique albums, got %d", count)
	}
	if _, err := db.Exec("INSERT INTO album(path) VALUES (?)", path); err == nil {
		t.Fatal("expected duplicate path error")
	}
	displayAlbums(db)
}

func TestInsertAlbumsHasCover(t *testing.T) {
	db := newAlbumTestDB(t)
	root := t.TempDir()
	albums := []album.Album{
		{Path: filepath.Join(root, "JPG")},
		{Path: filepath.Join(root, "WebP")},
		{Path: filepath.Join(root, "InvalidExtension")},
		{Path: filepath.Join(root, "Missing")},
	}
	for _, a := range albums {
		if err := os.MkdirAll(a.Path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, cover := range []string{
		filepath.Join(albums[0].Path, "cover.jpg"),
		filepath.Join(albums[1].Path, "cover.webp"),
		filepath.Join(albums[2].Path, "cover.png"),
	} {
		if err := os.WriteFile(cover, []byte("cover"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := insertAlbums(db, albums); err != nil {
		t.Fatal(err)
	}
	for i, want := range []bool{true, true, false, false} {
		var got bool
		if err := db.QueryRow("SELECT hascover FROM album WHERE path = ?", albums[i].Path).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("hascover for %s = %t, want %t", albums[i].Path, got, want)
		}
	}
}

func TestInsertAlbumsMissingArtist(t *testing.T) {
	db := newAlbumTestDB(t)
	if err := insertAlbums(db, []album.Album{{Path: "/missing/Album"}}); err != nil {
		t.Fatal(err)
	}
	var id sql.NullInt64
	if err := db.QueryRow("SELECT artistid FROM album").Scan(&id); err != nil {
		t.Fatal(err)
	}
	if id.Valid {
		t.Fatalf("expected NULL artistid, got %v", id)
	}
	displayAlbums(db)
}

func TestInsertAlbumsEmptyAndErrors(t *testing.T) {
	db := newAlbumTestDB(t)
	if err := insertAlbums(db, nil); err != nil {
		t.Fatal(err)
	}
	displayAlbums(db)
	if err := insertAlbums(db, []album.Album{{}}); err == nil || !strings.Contains(err.Error(), "missing a path") {
		t.Fatalf("expected missing path error, got %v", err)
	}
	if err := createTableAlbum(db); err == nil {
		t.Fatal("expected duplicate table error")
	}
	var hasCover bool
	if _, err := db.Exec("INSERT INTO album(path) VALUES (?)", "/default-cover"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT hascover FROM album WHERE path = ?", "/default-cover").Scan(&hasCover); err != nil {
		t.Fatal(err)
	}
	if hasCover {
		t.Fatal("hascover must default to false")
	}
	if _, err := db.Exec("DROP TABLE album"); err != nil {
		t.Fatal(err)
	}
	if err := insertAlbums(db, []album.Album{{Path: "/missing/Album"}}); err == nil {
		t.Fatal("expected insert error")
	}
}
