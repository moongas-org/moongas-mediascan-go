package main

import (
	"testing"
)

func TestMediaFileTableOperations(t *testing.T) {
	db := newTestDB(t)
	createTableMediaFile(db)

	insertMediaFile(db, "/music/Artist/Album/song.mp3", "/music/Artist/Album", "/music/Artist",
		"Song", "Artist", "Album Artist", "Album", "Rock", 2026)

	var id, year int
	var path, albumPath, artistPath, title, artist, albumArtist, album, genre string
	if err := db.QueryRow("SELECT * FROM mediafile").
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
