package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
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
	if len(trackColumns) != 16 || len(mediaColumns) != 16 {
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
	hasMediaColumn := trackColumns[13]
	if hasMediaColumn.Name != "hasmedia" || hasMediaColumn.Type != "BOOLEAN" ||
		hasMediaColumn.NotNull != 1 || !hasMediaColumn.Default.Valid || hasMediaColumn.PrimaryKey != 0 {
		t.Fatalf("unexpected hasmedia column: %+v", hasMediaColumn)
	}
	if _, err := db.Exec("INSERT INTO track(path) VALUES (?)", "default.yml"); err != nil {
		t.Fatal(err)
	}
	var hasSpotify, hasMedia bool
	if err := db.QueryRow("SELECT hasspotify, hasmedia FROM track WHERE path = ?", "default.yml").Scan(&hasSpotify, &hasMedia); err != nil {
		t.Fatal(err)
	}
	if hasSpotify {
		t.Fatal("hasspotify must default to false")
	}
	if hasMedia {
		t.Fatal("hasmedia must default to false")
	}
	if _, err := db.Exec("INSERT INTO track(path, hasmedia) VALUES (?, NULL)", "null.yml"); err == nil {
		t.Fatal("hasmedia must not allow NULL")
	}
	for i, name := range []string{"haslyricstxt", "haslyricslrc"} {
		c := trackColumns[14+i]
		if c.Name != name || c.Type != "BOOLEAN" || c.NotNull != 1 || !c.Default.Valid || c.PrimaryKey != 0 {
			t.Fatalf("unexpected %s column: %+v", name, c)
		}
		var got bool
		if err := db.QueryRow("SELECT "+name+" FROM track WHERE path = ?", "default.yml").Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got {
			t.Errorf("%s must default to false", name)
		}
		if _, err := db.Exec("INSERT INTO track(path, "+name+") VALUES (?, NULL)", name+".yml"); err == nil {
			t.Errorf("%s must not allow NULL", name)
		}
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
	if err := insertTracks(db, tracks, nil); err != nil {
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
	if err := insertTracks(db, tracks, nil); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("expected duplicate path error identifying track, got %v", err)
	}
}

func TestInsertTracksDetectsLyrics(t *testing.T) {
	for _, ext := range []string{".txt", ".lrc"} {
		for _, scenario := range []struct {
			name string
			kind string
			want bool
		}{
			{"missing", "missing", false},
			{"regular-file", "file", true},
			{"both-formats", "both", true},
			{"different-basename", "other", false},
			{"different-directory", "nested", false},
			{"uppercase-extension", "uppercase", false},
			{"directory", "directory", false},
			{"symlink-to-file", "symlink", true},
			{"broken-symlink", "broken", false},
			{"symlink-to-directory", "directory-link", false},
			{"access-error", "loop", false},
		} {
			t.Run(ext+"/"+scenario.name, func(t *testing.T) {
				db := newTestDB(t)
				createTableArtist(db)
				if err := createTableAlbum(db); err != nil {
					t.Fatal(err)
				}
				if err := createTableTrack(db); err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				path := filepath.Join(dir, "song.live.yml")
				lyricsPath := filepath.Join(dir, "song.live"+ext)
				write := func(path string) {
					t.Helper()
					if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, nil, 0o644); err != nil {
						t.Fatal(err)
					}
				}
				link := func(target string) {
					t.Helper()
					if err := os.Symlink(target, lyricsPath); err != nil {
						t.Fatal(err)
					}
				}
				switch scenario.kind {
				case "file":
					write(lyricsPath)
				case "both":
					write(filepath.Join(dir, "song.live.txt"))
					write(filepath.Join(dir, "song.live.lrc"))
				case "other":
					write(filepath.Join(dir, "other"+ext))
				case "nested":
					write(filepath.Join(dir, "nested", "song.live"+ext))
				case "uppercase":
					write(filepath.Join(dir, "song.live"+strings.ToUpper(ext)))
				case "directory":
					if err := os.Mkdir(lyricsPath, 0o755); err != nil {
						t.Fatal(err)
					}
				case "symlink":
					write(filepath.Join(dir, "source"))
					link("source")
				case "broken":
					link("missing")
				case "directory-link":
					link(dir)
				case "loop":
					link(filepath.Base(lyricsPath))
				}
				var logs bytes.Buffer
				previous := log.Writer()
				log.SetOutput(&logs)
				t.Cleanup(func() { log.SetOutput(previous) })
				if err := insertTracks(db, []track.Track{{Path: path}}, nil); err != nil {
					t.Fatal(err)
				}
				var txt, lrc bool
				if err := db.QueryRow("SELECT haslyricstxt, haslyricslrc FROM track WHERE path = ?", path).Scan(&txt, &lrc); err != nil {
					t.Fatal(err)
				}
				wantTxt, wantLrc := ext == ".txt" && scenario.want, ext == ".lrc" && scenario.want
				if scenario.kind == "both" {
					wantTxt, wantLrc = true, true
				}
				if txt != wantTxt || lrc != wantLrc {
					t.Errorf("lyrics flags = (%t, %t), want (%t, %t)", txt, lrc, wantTxt, wantLrc)
				}
				if err := displayTracks(db); err != nil {
					t.Fatal(err)
				}
				for _, expected := range []string{
					fmt.Sprintf("haslyricstxt= %t", wantTxt),
					fmt.Sprintf("haslyricslrc= %t", wantLrc),
				} {
					if !strings.Contains(logs.String(), expected) {
						t.Errorf("missing %q in track display", expected)
					}
				}
				if scenario.kind == "loop" && !strings.Contains(logs.String(), "ERROR accessing "+lyricsPath) {
					t.Fatalf("expected lyrics access error in logs, got %s", logs.String())
				}
			})
		}
	}
}

func TestInsertTracksDetectsMedia(t *testing.T) {
	tests := []struct {
		name       string
		mediaFiles []string
		mediaDirs  []string
		linkTarget string
		extensions []string
		want       bool
	}{
		{name: "no-media", extensions: []string{".mp3"}},
		{name: "matching-media", mediaFiles: []string{"song.live.mp3"}, extensions: []string{".mp3"}, want: true},
		{name: "multiple-media", mediaFiles: []string{"song.live.mp3", "song.live.m4a"}, extensions: []string{".mp3", ".m4a"}, want: true},
		{name: "second-extension", mediaFiles: []string{"song.live.m4a"}, extensions: []string{".mp3", ".m4a"}, want: true},
		{name: "unconfigured-extension", mediaFiles: []string{"song.live.flac"}, extensions: []string{".mp3"}},
		{name: "no-configured-extensions", mediaFiles: []string{"song.live.mp3"}},
		{name: "different-basename", mediaFiles: []string{"other.mp3", "song.live.extra.mp3"}, extensions: []string{".mp3"}},
		{name: "different-directory", mediaFiles: []string{"nested/song.live.mp3"}, extensions: []string{".mp3"}},
		{name: "case-sensitive-extension", mediaFiles: []string{"song.live.MP3"}, extensions: []string{".mp3"}},
		{name: "configured-uppercase-extension", mediaFiles: []string{"song.live.MP3"}, extensions: []string{".MP3"}, want: true},
		{name: "directory-not-media", mediaDirs: []string{"song.live.mp3"}, extensions: []string{".mp3"}},
		{name: "symlink-to-file", mediaFiles: []string{"source.bin"}, linkTarget: "source.bin", extensions: []string{".mp3"}, want: true},
		{name: "broken-symlink", linkTarget: "missing.mp3", extensions: []string{".mp3"}},
		{name: "symlink-to-directory", mediaDirs: []string{"source"}, linkTarget: "source", extensions: []string{".mp3"}},
		{name: "access-error-continues", mediaFiles: []string{"song.live.m4a"}, linkTarget: "song.live.mp3", extensions: []string{".mp3", ".m4a"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			createTableArtist(db)
			if err := createTableAlbum(db); err != nil {
				t.Fatal(err)
			}
			if err := createTableTrack(db); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			for _, name := range tt.mediaFiles {
				path := filepath.Join(dir, name)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range tt.mediaDirs {
				if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if tt.linkTarget != "" {
				if err := os.Symlink(tt.linkTarget, filepath.Join(dir, "song.live.mp3")); err != nil {
					t.Fatal(err)
				}
			}
			var logs bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(previous) })
			path := filepath.Join(dir, "song.live.yml")
			if err := insertTracks(db, []track.Track{{Path: path}}, tt.extensions); err != nil {
				t.Fatal(err)
			}
			var got bool
			if err := db.QueryRow("SELECT hasmedia FROM track WHERE path = ?", path).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("hasmedia = %t, want %t", got, tt.want)
			}
			if tt.name == "access-error-continues" && !strings.Contains(logs.String(), "ERROR accessing "+filepath.Join(dir, "song.live.mp3")) {
				t.Fatalf("expected media access error in logs, got %s", logs.String())
			}
		})
	}
}

func TestScannedTracksHaveMedia(t *testing.T) {
	db := newTestDB(t)
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"with-media", "without-media", "unconfigured-media"} {
		if err := os.WriteFile(filepath.Join(dir, name+".yml"), []byte(`trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: Album
  genre: Rock
  year: 2026
`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"with-media.mp3", "unconfigured-media.flac"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := scanIntoDB(db, conf.MediascanConf{MediaDirs: []string{root}, MediaExts: []string{".mp3"}}); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{
		"with-media": true, "without-media": false, "unconfigured-media": false,
	} {
		var got bool
		if err := db.QueryRow("SELECT hasmedia FROM track WHERE path = ?", filepath.Join(dir, name+".yml")).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s: hasmedia = %t, want %t", name, got, want)
		}
	}
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previous) })
	if err := displayTracks(db); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"hasmedia= true", "hasmedia= false"} {
		if !strings.Contains(logs.String(), expected) {
			t.Errorf("missing %q in track display", expected)
		}
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
	if err := insertTracks(db, nil, nil); err != nil {
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
	if err := insertTracks(db, []track.Track{{Path: "song.yml"}}, nil); err == nil {
		t.Fatal("expected insert error on closed database")
	}
	if err := displayTracks(db); err == nil {
		t.Fatal("expected query error on closed database")
	}
}
