package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
)

func createTableTrackStreaming(db *sql.DB) error {
	log.Println("Create trackstreaming table...")
	_, err := db.Exec(`CREATE TABLE trackstreaming (
		"trackid" INTEGER NOT NULL REFERENCES track(id),
		"streamingservice" TEXT NOT NULL,
		"streamingtrackid" TEXT NOT NULL,
		"streamingtrackurl" TEXT NOT NULL
	  );`)
	if err != nil {
		return fmt.Errorf("create trackstreaming table: %w", err)
	}
	if _, err := db.Exec("CREATE INDEX idx_trackstreaming_trackid ON trackstreaming (trackid)"); err != nil {
		return fmt.Errorf("create trackstreaming index: %w", err)
	}
	log.Println("trackstreaming table created")
	return nil
}

func insertTrackStreaming(db *sql.DB, tracks []track.Track) error {
	for _, t := range tracks {
		if t.TrackYamlFile.TrackIDData == nil || len(t.TrackYamlFile.TrackIDData.StreamingData) == 0 {
			continue
		}
		var trackID int64
		if err := db.QueryRow("SELECT id FROM track WHERE path = ?", t.Path).Scan(&trackID); err != nil {
			return fmt.Errorf("find track ID for %s: %w", t.Path, err)
		}
		for _, streaming := range t.TrackYamlFile.TrackIDData.StreamingData {
			log.Println("Inserting trackstreaming record ...")
			if _, err := db.Exec(`INSERT INTO trackstreaming(trackid, streamingservice, streamingtrackid, streamingtrackurl)
				VALUES (?, ?, ?, ?)`, trackID, string(streaming.StreamingService), streaming.TrackID, streaming.TrackURL); err != nil {
				return fmt.Errorf("insert streaming service %s for track %s: %w", streaming.StreamingService, t.Path, err)
			}
		}
	}
	return nil
}

func displayTrackStreaming(db *sql.DB) error {
	rows, err := db.Query("SELECT trackid, streamingservice, streamingtrackid, streamingtrackurl FROM trackstreaming ORDER BY trackid, streamingservice")
	if err != nil {
		return fmt.Errorf("query trackstreaming: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var trackID int64
		var service, streamingTrackID, streamingTrackURL string
		if err := rows.Scan(&trackID, &service, &streamingTrackID, &streamingTrackURL); err != nil {
			return fmt.Errorf("scan trackstreaming row: %w", err)
		}
		log.Println("trackstreaming: ", trackID, " ", service, " ", streamingTrackID, " ", streamingTrackURL)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read trackstreaming: %w", err)
	}
	return nil
}
