package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	track_scanner "github.com/moongas-org/moongas-mediascan-go/internal/shared/scanner/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/conf"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track/streaming"
)

func TestInsertTrackStreamingCorrelatesTrackIDs(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrackStreaming(db); err != nil {
		t.Fatal(err)
	}
	if err := insertTracks(db, []track.Track{{Path: "unrelated.yml"}}); err != nil {
		t.Fatal(err)
	}
	tracks := []track.Track{
		{
			Path: "first.yml",
			TrackYamlFile: track.TrackYamlFile{
				TrackIDData: &track.TrackIDData{
					StreamingData: []streaming.StreamingData[any]{
						{StreamingService: streaming.Spotify, TrackID: "07RXBKfyCYIYRMLCvlGWXU"},
						{StreamingService: streaming.AppleMusic, TrackID: "apple-id"},
					},
				},
			},
		},
		{Path: "no-streaming.yml"},
		{Path: "empty-streaming.yml", TrackYamlFile: track.TrackYamlFile{TrackIDData: &track.TrackIDData{}}},
		{
			Path: "second.yml",
			TrackYamlFile: track.TrackYamlFile{
				TrackIDData: &track.TrackIDData{
					StreamingData: []streaming.StreamingData[any]{
						{StreamingService: streaming.Spotify, TrackID: "second-id"},
					},
				},
			},
		},
	}
	if err := insertTracks(db, tracks); err != nil {
		t.Fatal(err)
	}
	// Reverse the list to verify IDs are resolved by path, not slice position.
	for i, j := 0, len(tracks)-1; i < j; i, j = i+1, j-1 {
		tracks[i], tracks[j] = tracks[j], tracks[i]
	}
	if err := insertTrackStreaming(db, tracks); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT t.path, s.streamingservice, s.streamingtrackid
		FROM trackstreaming s JOIN track t ON t.id = s.trackid ORDER BY t.path, s.streamingservice`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got [][3]string
	for rows.Next() {
		var row [3]string
		if err := rows.Scan(&row[0], &row[1], &row[2]); err != nil {
			t.Fatal(err)
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := [][3]string{
		{"first.yml", "apple-music", "apple-id"},
		{"first.yml", "spotify", "07RXBKfyCYIYRMLCvlGWXU"},
		{"second.yml", "spotify", "second-id"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("streaming rows = %v, want %v", got, want)
	}
	if err := displayTrackStreaming(db); err != nil {
		t.Fatal(err)
	}
}

func TestInsertTrackStreamingFromSingleScan(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrackStreaming(db); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	dir := filepath.Join(root, "Artist", "Album")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	input := `trackData:
  title: Song
  artist: Artist
  albumartist: Artist
  album: Album
  genre: Rock
  year: 2026
trackIdData:
  streamingData:
    - streamingService: spotify
      trackUrl: https://open.spotify.com/track/07RXBKfyCYIYRMLCvlGWXU
      trackId: 07RXBKfyCYIYRMLCvlGWXU
      apiData: null
`
	path := filepath.Join(dir, "song.yml")
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	tracks := track_scanner.ScanTracks(conf.MediascanConf{MediaDirs: []string{root}})
	if len(tracks) != 1 {
		t.Fatalf("expected one scanned track, got %v", tracks)
	}
	if err := insertTracks(db, tracks); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := insertTrackStreaming(db, tracks); err != nil {
		t.Fatal(err)
	}
	var gotPath, service, streamingID string
	if err := db.QueryRow(`SELECT t.path, s.streamingservice, s.streamingtrackid
		FROM trackstreaming s JOIN track t ON t.id = s.trackid`).
		Scan(&gotPath, &service, &streamingID); err != nil {
		t.Fatal(err)
	}
	if gotPath != path || service != "spotify" || streamingID != "07RXBKfyCYIYRMLCvlGWXU" {
		t.Fatalf("unexpected streaming row: %q %q %q", gotPath, service, streamingID)
	}
}

func TestTrackStreamingEmptyAndErrors(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrackStreaming(db); err != nil {
		t.Fatal(err)
	}
	if err := insertTrackStreaming(db, nil); err != nil {
		t.Fatal(err)
	}
	if err := displayTrackStreaming(db); err != nil {
		t.Fatal(err)
	}
	missing := []track.Track{{
		Path: "missing.yml",
		TrackYamlFile: track.TrackYamlFile{TrackIDData: &track.TrackIDData{
			StreamingData: []streaming.StreamingData[any]{{StreamingService: streaming.Spotify, TrackID: "id"}},
		}},
	}}
	if err := insertTrackStreaming(db, missing); err == nil || !strings.Contains(err.Error(), "missing.yml") {
		t.Fatalf("expected missing track error, got %v", err)
	}
	if err := createTableTrackStreaming(db); err == nil {
		t.Fatal("expected duplicate table error")
	}
	if err := insertTracks(db, missing); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DROP TABLE trackstreaming"); err != nil {
		t.Fatal(err)
	}
	if err := insertTrackStreaming(db, missing); err == nil {
		t.Fatal("expected streaming insert error")
	}
	if err := displayTrackStreaming(db); err == nil {
		t.Fatal("expected query error")
	}
}
