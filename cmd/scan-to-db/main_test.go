package main

import (
	"bytes"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
)

func captureSlog(t *testing.T, output *bytes.Buffer) {
	t.Helper()
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })
}

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
	captureSlog(t, &logs)
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3", ".flac"}}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{mp3Path, flacPath} {
		var gotArtistPath, gotAlbumPath, gotTrackPath, service string
		var hasArtistYaml, hasAlbumYaml, hasSpotify bool
		err := db.QueryRow(`SELECT a.path, a.hasartistyaml, b.path, b.hasalbumyaml, t.path, s.streamingservice, t.hasspotify
				FROM mediafile m
				JOIN artist a ON a.id = m.artistid
				JOIN album b ON b.id = m.albumid AND b.artistid = a.id
				JOIN track t ON t.id = m.trackid AND t.artistid = a.id AND t.albumid = b.id
				JOIN trackstreaming s ON s.trackid = t.id WHERE m.path = ?`, path).
			Scan(&gotArtistPath, &hasArtistYaml, &gotAlbumPath, &hasAlbumYaml, &gotTrackPath, &service, &hasSpotify)
		if err != nil {
			t.Fatal(err)
		}
		if gotArtistPath != artistPath || !hasArtistYaml || gotAlbumPath != albumPath || !hasAlbumYaml ||
			gotTrackPath != trackPath || service != "spotify" || !hasSpotify {
			t.Fatalf("incorrect relationships for %s", path)
		}
	}
	if err := displayTracks(db); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(logs.String(), "has_spotify=true") {
		t.Fatal("displayed track must include hasspotify")
	}
	last := -1
	for _, phase := range []string{
		"Scanning artist media dir", "Inserting artist record",
		"Scanning album media dir", "Inserting album record",
		"Scanning track media dir", "Inserting track record",
		"Inserting track streaming record", "Scanning file media dir", "Inserting mediafile record",
	} {
		index := strings.Index(logs.String(), phase)
		if index <= last {
			t.Fatalf("missing or out-of-order phase %q:\n%s", phase, logs.String())
		}
		last = index
	}
	if strings.Count(logs.String(), "Scanning track media dir") != 1 {
		t.Fatal("ScanTracks must run once")
	}
}

func TestScanIntoDBWithoutArtistOrAlbumYaml(t *testing.T) {
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
	captureSlog(t, &logs)
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}}); err != nil {
		t.Fatal(err)
	}
	var fallbackArtistID int64
	var name string
	var hasArtistYaml bool
	if err := db.QueryRow("SELECT id, name, hasartistyaml FROM artist WHERE path = ?", filepath.Dir(dir)).
		Scan(&fallbackArtistID, &name, &hasArtistYaml); err != nil {
		t.Fatal(err)
	}
	if name != "MissingArtist" || hasArtistYaml {
		t.Fatalf("unexpected fallback artist: name=%q hasartistyaml=%t", name, hasArtistYaml)
	}
	var fallbackAlbumID, albumArtistID int64
	var hasAlbumYaml bool
	if err := db.QueryRow("SELECT id, name, artistid, hasalbumyaml FROM album WHERE path = ?", dir).
		Scan(&fallbackAlbumID, &name, &albumArtistID, &hasAlbumYaml); err != nil {
		t.Fatal(err)
	}
	if name != "MissingAlbum" || albumArtistID != fallbackArtistID || hasAlbumYaml {
		t.Fatalf("unexpected fallback album: name=%q artistid=%d hasalbumyaml=%t", name, albumArtistID, hasAlbumYaml)
	}
	var artistID, albumID, trackID sql.NullInt64
	if err := db.QueryRow("SELECT artistid, albumid FROM track").Scan(&artistID, &albumID); err != nil {
		t.Fatal(err)
	}
	if !artistID.Valid || artistID.Int64 != fallbackArtistID || !albumID.Valid || albumID.Int64 != fallbackAlbumID {
		t.Fatal("track must link to the fallback artist and album")
	}
	if err := db.QueryRow("SELECT artistid, albumid, trackid FROM mediafile").Scan(&artistID, &albumID, &trackID); err != nil {
		t.Fatal(err)
	}
	if !artistID.Valid || artistID.Int64 != fallbackArtistID || !albumID.Valid || albumID.Int64 != fallbackAlbumID || trackID.Valid {
		t.Fatal("media file must link to the fallback artist and album, but not a track")
	}
	if !strings.Contains(logs.String(), "storing NULL ID") {
		t.Error("missing warning for unlinked media file")
	}
	var albums int
	if err := db.QueryRow("SELECT COUNT(*) FROM album").Scan(&albums); err != nil {
		t.Fatal(err)
	}
	if albums != 1 {
		t.Fatalf("got %d album rows, want one fallback album", albums)
	}
}
