package main

import "testing"

func TestArtistTableOperations(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	if err := insertArtist(db, "/music/Artist", "Artist", "City", "US", "TX", "en", true); err != nil {
		t.Fatal(err)
	}

	var id int
	var path, name, city, country, region, language string
	var hasArtistYaml bool
	if err := db.QueryRow("SELECT * FROM artist").
		Scan(&id, &path, &name, &city, &country, &region, &language, &hasArtistYaml); err != nil {
		t.Fatal(err)
	}
	if id != 1 || path != "/music/Artist" || name != "Artist" || city != "City" ||
		country != "US" || region != "TX" || language != "en" || !hasArtistYaml {
		t.Fatal("artist row differs from inserted data")
	}
	displayArtists(db)
}
