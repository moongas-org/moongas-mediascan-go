package main

import (
	"database/sql"
	"fmt"
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
	if len(trackColumns) != 13 || len(mediaColumns) != 16 {
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
	hasSpotifyColumn := trackColumns[12]
	if hasSpotifyColumn.Name != "hasspotify" || hasSpotifyColumn.Type != "BOOLEAN" ||
		hasSpotifyColumn.NotNull != 1 || !hasSpotifyColumn.Default.Valid || hasSpotifyColumn.PrimaryKey != 0 {
		t.Fatalf("unexpected hasspotify column: %+v", hasSpotifyColumn)
	}
	if _, err := db.Exec("INSERT INTO track(path) VALUES (?)", "default.yml"); err != nil {
		t.Fatal(err)
	}
	var hasSpotify bool
	if err := db.QueryRow("SELECT hasspotify FROM track WHERE path = ?", "default.yml").Scan(&hasSpotify); err != nil {
		t.Fatal(err)
	}
	if hasSpotify {
		t.Fatal("hasspotify must default to false")
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

func TestScannedTracksHaveSpotify(t *testing.T) {
	db := newTestDB(t)
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	entry := func(service, id string) string {
		return fmt.Sprintf(`    - streamingService: %q
      trackUrl: ""
      trackId: %q
      apiData: null
`, service, id)
	}
	streamingData := func(entries ...string) string {
		return "trackIdData:\n  streamingData:\n" + strings.Join(entries, "")
	}
	tests := []struct {
		name        string
		trackIDData string
		want        bool
	}{
		{"missing-track-id-data", "", false},
		{"null-track-id-data", "trackIdData: null\n", false},
		{"empty-track-id-data", "trackIdData: {}\n", false},
		{"null-streaming-data", "trackIdData:\n  streamingData: null\n", false},
		{"empty-streaming-data", "trackIdData:\n  streamingData: []\n", false},
		{"empty-spotify-id", streamingData(entry("spotify", "")), false},
		{"null-spotify-id", `trackIdData:
  streamingData:
    - {streamingService: spotify, trackUrl: "", trackId: null, apiData: null}
`, false},
		{"url-and-api-without-id", `trackIdData:
  streamingData:
    - {streamingService: spotify, trackUrl: "https://open.spotify.com/track/example", trackId: "", apiData: {id: example}}
`, false},
		{"other-service", streamingData(entry("apple-music", "apple-id")), false},
		{"id-on-different-service", streamingData(entry("spotify", ""), entry("apple-music", "apple-id")), false},
		{"service-case-differs", streamingData(entry("Spotify", "spotify-id")), false},
		{"spotify-id", streamingData(entry("spotify", "spotify-id")), true},
		{"spotify-after-other-service", streamingData(entry("apple-music", "apple-id"), entry("spotify", "spotify-id")), true},
		{"spotify-after-empty-id", streamingData(entry("spotify", ""), entry("spotify", "spotify-id")), true},
		{"spotify-before-empty-id", streamingData(entry("spotify", "spotify-id"), entry("spotify", "")), true},
		{"spotify-before-other-service", streamingData(entry("spotify", "spotify-id"), entry("apple-music", "")), true},
		{"nonempty-whitespace-id", streamingData(entry("spotify", " ")), true},
	}
	const core = `trackData:
  title: YAML Song
  artist: YAML Artist
  albumartist: YAML Album Artist
  album: YAML Album
  genre: Rock
  year: 2026
`
	for _, tt := range tests {
		path := filepath.Join(dir, tt.name+".yml")
		if err := os.WriteFile(path, []byte(core+tt.trackIDData), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM track").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(tests) {
		t.Fatalf("scanned %d tracks, want %d", count, len(tests))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			path := filepath.Join(dir, tt.name+".yml")
			if err := db.QueryRow("SELECT hasspotify FROM track WHERE path = ?", path).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("hasspotify = %t, want %t", got, tt.want)
			}
		})
	}
	var streamingID string
	if err := db.QueryRow(`SELECT s.streamingtrackid FROM trackstreaming s
		JOIN track t ON t.id = s.trackid WHERE t.path = ? AND s.streamingservice = ?`,
		filepath.Join(dir, "url-and-api-without-id.yml"), "spotify").Scan(&streamingID); err != nil {
		t.Fatal(err)
	}
	if streamingID != "" {
		t.Fatalf("streaming track ID = %q, want empty string", streamingID)
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
