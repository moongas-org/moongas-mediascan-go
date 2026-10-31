package main

import (
	"bytes"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
)

func newTestDB(t *testing.T) *sql.DB {
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
	return db
}

func TestScanIntoDBRelationshipsAndOrder(t *testing.T) {
	db := newTestDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	artistPath := filepath.Join(root, "Artist")
	albumPath := filepath.Join(artistPath, "Album")
	if err := os.MkdirAll(albumPath, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(artistPath, "artist.yml"), `artistData:
  artistNames: [Artist]
  languageCodes: [en]
`)
	write(filepath.Join(albumPath, "album.yml"), `albumData:
  name: Album
  dateReleased: {y: 2001}
  explicitCover: true
`)
	mp3Path := filepath.Join(albumPath, "different-name.mp3")
	flacPath := filepath.Join(albumPath, "another-name.flac")
	trackPath := filepath.Join(albumPath, "song.yml")
	write(trackPath, `trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: Album
  genre: Rock
  year: 2001
trackIdData:
  streamingData:
  - streamingService: spotify
    trackUrl: https://open.spotify.com/track/example
    trackId: example
    apiData: null
mediaFileData:
  - path: different-name.mp3
    modTime: 2024-09-16 00:34:15.638511+00:00
  - path: `+flacPath+`
`)
	write(mp3Path, "unreadable audio tags")
	write(flacPath, "unreadable audio tags")
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3", ".flac"}}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{mp3Path, flacPath} {
		var gotArtistPath, gotAlbumPath, gotTrackPath, service string
		err := db.QueryRow(`SELECT a.path, b.path, t.path, s.streamingservice
				FROM mediafile m
				JOIN artist a ON a.id = m.artistid
				JOIN album b ON b.id = m.albumid AND b.artistid = a.id
				JOIN track t ON t.id = m.trackid AND t.artistid = a.id AND t.albumid = b.id
				JOIN trackstreaming s ON s.trackid = t.id WHERE m.path = ?`, path).
			Scan(&gotArtistPath, &gotAlbumPath, &gotTrackPath, &service)
		if err != nil {
			t.Fatal(err)
		}
		if gotArtistPath != artistPath || gotAlbumPath != albumPath || gotTrackPath != trackPath || service != "spotify" {
			t.Fatalf("incorrect relationships for %s", path)
		}
	}
	last := -1
	for _, phase := range []string{
		"Scanning artist media dir:", "Inserting artist record",
		"Scanning album media dir:", "Inserting album record",
		"Scanning track media dir:", "Inserting track record",
		"Inserting trackstreaming record", "Scanning file media dir:", "Inserting mediafile record",
	} {
		index := strings.Index(logs.String(), phase)
		if index <= last {
			t.Fatalf("missing or out-of-order phase %q:\n%s", phase, logs.String())
		}
		last = index
	}
	if strings.Count(logs.String(), "Scanning track media dir:") != 1 {
		t.Fatal("ScanTracks must run once")
	}
}

func TestScanIntoDBMissingRelationships(t *testing.T) {
	db := newTestDB(t)
	root := t.TempDir()
	dir := filepath.Join(root, "MissingArtist", "MissingAlbum")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "song.yml"), []byte(`trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: Album
  genre: Rock
  year: 2001
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "song.mp3"), []byte("audio"), 0o644); err != nil {
		t.Fatal(err)
	}
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}}); err != nil {
		t.Fatal(err)
	}
	var artistID, albumID, trackID sql.NullInt64
	if err := db.QueryRow("SELECT artistid, albumid FROM track").Scan(&artistID, &albumID); err != nil {
		t.Fatal(err)
	}
	if artistID.Valid || albumID.Valid {
		t.Fatal("missing track parents must be NULL")
	}
	if err := db.QueryRow("SELECT artistid, albumid, trackid FROM mediafile").Scan(&artistID, &albumID, &trackID); err != nil {
		t.Fatal(err)
	}
	if artistID.Valid || albumID.Valid || trackID.Valid {
		t.Fatal("missing media-file relationships must be NULL, not matched by filename")
	}
	for _, warning := range []string{"WARNING: no artist row", "WARNING: no album row", "storing NULL trackid"} {
		if !strings.Contains(logs.String(), warning) {
			t.Errorf("missing warning %q", warning)
		}
	}
	var albums int
	if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&albums); err != nil {
		t.Fatal(err)
	}
	if albums != 0 {
		t.Fatal("must not create fallback albums from media files")
	}
}
