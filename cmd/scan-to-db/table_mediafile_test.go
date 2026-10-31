package main

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/media_file"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
)

func TestMediaFileTableOperations(t *testing.T) {
	db := newTestDB(t)
	createTableMediaFile(db)

	if err := insertMediaFile(db, "/music/Artist/Album/song.mp3", "/music/Artist/Album", "/music/Artist",
		"Song", "Artist", "Album Artist", "Album", "Rock", 2026, sql.NullInt64{}, sql.NullInt64{}, sql.NullInt64{}); err != nil {
		t.Fatal(err)
	}

	var id, year int
	var path, albumPath, artistPath, title, artist, albumArtist, album, genre string
	if err := db.QueryRow("SELECT id, path, albumpath, artistpath, title, artist, albumartist, album, genre, year FROM mediafile").
		Scan(&id, &path, &albumPath, &artistPath, &title, &artist, &albumArtist, &album, &genre, &year); err != nil {
		t.Fatal(err)
	}
	if id != 1 || path != "/music/Artist/Album/song.mp3" || albumPath != "/music/Artist/Album" ||
		artistPath != "/music/Artist" || title != "Song" || artist != "Artist" ||
		albumArtist != "Album Artist" || album != "Album" || genre != "Rock" || year != 2026 {
		t.Fatal("mediafile row differs from inserted data")
	}

	for _, index := range []string{"idx_artistpath", "idx_albumpath"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?", index).
			Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("missing mediafile index %s", index)
		}
	}

	displayMediaFiles(db)
}

func TestInsertMediaFilesUsesActualRowIDs(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := createTableAlbum(db); err != nil {
		t.Fatal(err)
	}
	if err := createTableTrack(db); err != nil {
		t.Fatal(err)
	}
	createTableMediaFile(db)
	for _, statement := range []string{
		`INSERT INTO artist(id, path) VALUES (7, '/music/Artist')`,
		`INSERT INTO album(id, path, artistid) VALUES (19, '/music/Artist/Album', 7)`,
		`INSERT INTO track(id, path) VALUES (31, '/music/Artist/Album/unrelated.yml')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	tracks := []track.Track{{
		Path: "/music/Artist/Album/song.yml",
		TrackYamlFile: track.TrackYamlFile{
			MediaFileData: []media_file.MediaFileData{
				{Path: utils.Pointer("./audio.mp3")},
				{},
			},
		},
	}}
	if err := insertTracks(db, tracks); err != nil {
		t.Fatal(err)
	}
	files := []media_file.MediaFileData{{
		Path:  utils.Pointer("/music/Artist/Album/audio.mp3"),
		Title: utils.Pointer("Song"), Artist: utils.Pointer("Artist"),
		AlbumArtist: utils.Pointer("Artist"), Album: utils.Pointer("Album"),
		Genre: utils.Pointer("Rock"), Year: utils.Pointer(2001),
	}}
	if err := insertMediaFiles(db, files, tracks); err != nil {
		t.Fatal(err)
	}
	var artistID, albumID, trackID int
	if err := db.QueryRow("SELECT artistid, albumid, trackid FROM mediafile").
		Scan(&artistID, &albumID, &trackID); err != nil {
		t.Fatal(err)
	}
	if artistID != 7 || albumID != 19 || trackID != 32 {
		t.Fatalf("incorrect IDs: %d %d %d", artistID, albumID, trackID)
	}
	if err := db.QueryRow("SELECT artistid, albumid FROM track WHERE id = 32").Scan(&artistID, &albumID); err != nil {
		t.Fatal(err)
	}
	if artistID != 7 || albumID != 19 {
		t.Fatalf("incorrect track parents: %d %d", artistID, albumID)
	}
	displayMediaFiles(db)
}

func TestInsertMediaFilesRejectsConflictingTracks(t *testing.T) {
	dir := t.TempDir()
	tracks := []track.Track{
		{Path: filepath.Join(dir, "first.yml"), TrackYamlFile: track.TrackYamlFile{
			MediaFileData: []media_file.MediaFileData{{Path: utils.Pointer("song.mp3")}},
		}},
		{Path: filepath.Join(dir, "second.yml"), TrackYamlFile: track.TrackYamlFile{
			MediaFileData: []media_file.MediaFileData{{Path: utils.Pointer(filepath.Join(dir, "song.mp3"))}},
		}},
	}
	err := insertMediaFiles(newTestDB(t), nil, tracks)
	if err == nil || !strings.Contains(err.Error(), "claimed by both") {
		t.Fatalf("expected conflicting track paths error, got %v", err)
	}
}

func TestInsertMediaFilesMissingPath(t *testing.T) {
	for _, path := range []*string{nil, utils.Pointer("")} {
		err := insertMediaFiles(newTestDB(t), []media_file.MediaFileData{{Path: path}}, nil)
		if err == nil || !strings.Contains(err.Error(), "missing a path") {
			t.Fatalf("expected missing path error, got %v", err)
		}
	}
}
