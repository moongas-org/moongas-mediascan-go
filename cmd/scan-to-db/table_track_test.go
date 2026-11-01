package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	track_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
)

func TestTrackTableSchemaMatchesMediaFile(t *testing.T) {
	db := newTestDB(t)
	createTableMediaFile(db)
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	type column struct {
		ID         int
		Name, Type string
		NotNull    int
		Default    sql.NullString
		PrimaryKey int
	}
	columns := func(table string) []column {
		t.Helper()
		rows, err := db.Query("PRAGMA table_info(" + table + ")")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var result []column
		for rows.Next() {
			var c column
			if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.NotNull, &c.Default, &c.PrimaryKey); err != nil {
				t.Fatal(err)
			}
			result = append(result, c)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return result
	}
	trackColumns, mediaColumns := columns("track"), columns("mediafile")
	if len(trackColumns) != 12 || len(mediaColumns) != 16 {
		t.Fatalf("unexpected schema sizes: track=%d mediafile=%d", len(trackColumns), len(mediaColumns))
	}
	for i := 0; i < 10; i++ {
		if trackColumns[i] != mediaColumns[i] {
			t.Errorf("track column %+v differs from mediafile column %+v", trackColumns[i], mediaColumns[i])
		}
	}
	if mediaColumns[10].Name != "size" || mediaColumns[10].Type != "INTEGER" ||
		mediaColumns[11].Name != "format" || mediaColumns[11].Type != "TEXT" ||
		mediaColumns[12].Name != "duration" || mediaColumns[12].Type != "REAL" {
		t.Fatal("mediafile size, format, or duration column has an unexpected schema")
	}
	if trackColumns[10].Name != "artistid" || trackColumns[11].Name != "albumid" ||
		mediaColumns[13].Name != "artistid" || mediaColumns[14].Name != "albumid" || mediaColumns[15].Name != "trackid" {
		t.Fatal("missing relationship columns")
	}
	for _, index := range []string{"idx_track_artistpath", "idx_track_albumpath"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND tbl_name = 'track' AND name = ?", index).
			Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("missing track index %s", index)
		}
	}
}

func TestInsertScannedTracks(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "song.yml")
	input := `trackData:
  title: YAML Song
  artist: YAML Artist
  albumartist: YAML Album Artist
  album: YAML Album
  genre: Rock
  year: 2026
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	tracks := track_scanner.ScanTracks(conf.MediascanConf{MediaDirs: []string{root}})
	if len(tracks) != 1 {
		t.Fatalf("expected one scanned track without an audio file, got %+v", tracks)
	}
	if err := insertTracks(db, tracks); err != nil {
		t.Fatal(err)
	}
	var id, year int
	var gotPath, albumPath, artistPath, title, artist, albumArtist, album, genre string
	if err := db.QueryRow("SELECT id, path, albumpath, artistpath, title, artist, albumartist, album, genre, year FROM track").
		Scan(&id, &gotPath, &albumPath, &artistPath, &title, &artist, &albumArtist, &album, &genre, &year); err != nil {
		t.Fatal(err)
	}
	if id != 1 || gotPath != path || albumPath != dir || artistPath != filepath.Dir(dir) ||
		title != "YAML Song" || artist != "YAML Artist" || albumArtist != "YAML Album Artist" ||
		album != "YAML Album" || genre != "Rock" || year != 2026 {
		t.Fatal("track row differs from scanned YAML metadata")
	}
	if err := displayTracks(db); err != nil {
		t.Fatal(err)
	}
	if err := insertTracks(db, tracks); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("expected duplicate path error identifying track, got %v", err)
	}
}

func TestTrackTableEmptyAndErrors(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	if err := insertTracks(db, nil); err != nil {
		t.Fatal(err)
	}
	if err := displayTracks(db); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM track").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected empty track table, got %d rows", count)
	}
	if err := createTableTrack(db); err == nil {
		t.Fatal("expected table creation error")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := insertTracks(db, []track.Track{{Path: "song.yml"}}); err == nil {
		t.Fatal("expected insert error on closed database")
	}
	if err := displayTracks(db); err == nil {
		t.Fatal("expected query error on closed database")
	}
}
