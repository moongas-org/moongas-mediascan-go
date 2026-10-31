package main

import (
	"database/sql"
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
	insertArtist(db, "/unrelated", "Other", "", "", "", "")
	insertArtist(db, artistPath, "Artist", "", "", "", "")
	path := filepath.Join(artistPath, "Album")
	albums := []album.Album{
		{Path: path, AlbumYamlFile: album.AlbumYamlFile{AlbumData: album.AlbumData{
			Name: "YAML album", DateReleased: date.Date{Y: 1993}, ExplicitCover: true,
		}}},
		{Path: path},
		{Path: filepath.Join(artistPath, "Second")},
	}
	if err := insertAlbums(db, albums); err != nil {
		t.Fatal(err)
	}
	var name string
	var year, artistID, count int
	var explicit bool
	if err := db.QueryRow("SELECT name, year, explicitcover, artistid FROM album WHERE path = ?", path).
		Scan(&name, &year, &explicit, &artistID); err != nil {
		t.Fatal(err)
	}
	if name != "YAML album" || year != 1993 || !explicit || artistID != 2 {
		t.Fatalf("unexpected album: %s %d %t artistid=%d", name, year, explicit, artistID)
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
	if _, err := db.Exec("DROP TABLE album"); err != nil {
		t.Fatal(err)
	}
	if err := insertAlbums(db, []album.Album{{Path: "/missing/Album"}}); err == nil {
		t.Fatal("expected insert error")
	}
}
