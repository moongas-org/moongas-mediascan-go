package main

import "testing"

func TestArtistTableOperations(t *testing.T) {
	db := newTestDB(t)
	createTableArtist(db)
	insertArtist(db, "/music/Artist", "Artist", "City", "US", "TX", "en")

	var id int
	var path, name, city, country, region, language string
	if err := db.QueryRow("SELECT * FROM artist").
		Scan(&id, &path, &name, &city, &country, &region, &language); err != nil {
		t.Fatal(err)
	}
	if id != 1 || path != "/music/Artist" || name != "Artist" || city != "City" ||
		country != "US" || region != "TX" || language != "en" {
		t.Fatal("artist row differs from inserted data")
	}
	displayArtists(db)
}
