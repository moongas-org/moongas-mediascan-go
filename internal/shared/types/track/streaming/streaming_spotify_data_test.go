package streaming

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStreamingSpotifyDataJSONRoundTrip(t *testing.T) {
	input := []byte(`{
		"streamingService": "spotify",
		"trackUrl": "https://open.spotify.com/track/0ZYf6N5eBeXStMri0mf3DZ",
		"trackId": "0ZYf6N5eBeXStMri0mf3DZ",
		"isrc": "USSM10027739",
		"apiData": {
			"tracks": {
				"href": "https://api.spotify.com/v1/search?offset=0&limit=5&query=isrc%3AUSSM10027739&type=track",
				"limit": 5,
				"next": null,
				"offset": 0,
				"previous": null,
				"total": 0,
				"items": [{
					"album": {
						"album_type": "album",
						"artists": [{
							"external_urls": {"spotify": "https://open.spotify.com/artist/1L0y9srZMyh9XUnYGv37IP"},
							"href": "https://api.spotify.com/v1/artists/1L0y9srZMyh9XUnYGv37IP",
							"id": "1L0y9srZMyh9XUnYGv37IP",
							"name": "'Til Tuesday",
							"type": "artist",
							"uri": "spotify:artist:1L0y9srZMyh9XUnYGv37IP"
						}],
						"external_urls": {"spotify": "https://open.spotify.com/album/1y4s0JN8CQMRwPTJ64jCUp"},
						"href": "https://api.spotify.com/v1/albums/1y4s0JN8CQMRwPTJ64jCUp",
						"id": "1y4s0JN8CQMRwPTJ64jCUp",
						"images": [
							{"height": 640, "width": 640, "url": "https://i.scdn.co/image/ab67616d0000b27322dc4d39d74c4c541f495103"},
							{"height": 300, "width": 300, "url": "https://i.scdn.co/image/ab67616d00001e0222dc4d39d74c4c541f495103"},
							{"height": 64, "width": 64, "url": "https://i.scdn.co/image/ab67616d0000485122dc4d39d74c4c541f495103"}
						],
						"is_playable": true,
						"name": "Voices Carry (Expanded Edition)",
						"release_date": "1985-06-01",
						"release_date_precision": "day",
						"total_tracks": 14,
						"type": "album",
						"uri": "spotify:album:1y4s0JN8CQMRwPTJ64jCUp"
					},
					"artists": [{
						"external_urls": {"spotify": "https://open.spotify.com/artist/1L0y9srZMyh9XUnYGv37IP"},
						"href": "https://api.spotify.com/v1/artists/1L0y9srZMyh9XUnYGv37IP",
						"id": "1L0y9srZMyh9XUnYGv37IP",
						"name": "'Til Tuesday",
						"type": "artist",
						"uri": "spotify:artist:1L0y9srZMyh9XUnYGv37IP"
					}],
					"disc_number": 1,
					"duration_ms": 263573,
					"explicit": false,
					"external_ids": {"isrc": "USSM10027739"},
					"external_urls": {"spotify": "https://open.spotify.com/track/0ZYf6N5eBeXStMri0mf3DZ"},
					"href": "https://api.spotify.com/v1/tracks/0ZYf6N5eBeXStMri0mf3DZ",
					"id": "0ZYf6N5eBeXStMri0mf3DZ",
					"is_local": false,
					"is_playable": true,
					"name": "Voices Carry",
					"track_number": 5,
					"type": "track",
					"uri": "spotify:track:0ZYf6N5eBeXStMri0mf3DZ"
				}]
			}
		}
	}`)

	var data StreamingSpotifyData
	if err := json.Unmarshal(input, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.StreamingService != Spotify || data.TrackID != "0ZYf6N5eBeXStMri0mf3DZ" ||
		data.TrackURL != "https://open.spotify.com/track/0ZYf6N5eBeXStMri0mf3DZ" ||
		data.ISRC == nil || *data.ISRC != "USSM10027739" {
		t.Fatalf("unexpected track identifiers: %+v", data.StreamingData)
	}
	page := data.APIData.Tracks
	if page.Limit != 5 || page.Total != 0 || page.Next != nil || page.Previous != nil || len(page.Items) != 1 {
		t.Fatalf("unexpected track page: %+v", page)
	}
	track := page.Items[0]
	if track.Name != "Voices Carry" || track.DurationMS != 263573 || track.ExternalIDs.ISRC != "USSM10027739" {
		t.Errorf("unexpected track: %+v", track)
	}
	if track.Album.Name != "Voices Carry (Expanded Edition)" ||
		len(track.Album.Images) != 3 || track.Album.Images[0].Height == nil || *track.Album.Images[0].Height != 640 {
		t.Errorf("unexpected album: %+v", track.Album)
	}

	out, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var want, got map[string]any
	if err := json.Unmarshal(input, &want); err != nil {
		t.Fatalf("unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal output JSON: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response changed during round trip:\ngot: %s\nwant: %s", out, input)
	}
	yamlOut, err := yaml.Marshal(data)
	if err != nil {
		t.Fatalf("marshal YAML: %v", err)
	}
	var yamlDecoded StreamingSpotifyData
	if err := yaml.Unmarshal(yamlOut, &yamlDecoded); err != nil {
		t.Fatalf("unmarshal YAML: %v", err)
	}
	if !reflect.DeepEqual(yamlDecoded, data) {
		t.Errorf("typed Spotify data changed during YAML round trip: %+v", yamlDecoded)
	}
	var yamlWant, yamlGot map[string]any
	if err := yaml.Unmarshal(input, &yamlWant); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(yamlOut, &yamlGot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(yamlGot, yamlWant) {
		t.Errorf("Spotify YAML keys changed:\ngot: %s\nwant: %s", yamlOut, input)
	}
}

func TestSpotifyAPIDataPaginationAndNullableImages(t *testing.T) {
	input := []byte(`{"tracks":{
		"next":"https://api.spotify.com/v1/search?offset=5",
		"previous":"https://api.spotify.com/v1/search?offset=0",
		"items":[
			{"id":"first","album":{"images":[{"height":null,"width":null,"url":"artwork"}]}},
			{"id":"second"}
		]
	}}`)
	var data SpotifyAPIData
	if err := json.Unmarshal(input, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	page := data.Tracks
	if page.Next == nil || *page.Next != "https://api.spotify.com/v1/search?offset=5" ||
		page.Previous == nil || *page.Previous != "https://api.spotify.com/v1/search?offset=0" {
		t.Errorf("unexpected pagination: %+v", page)
	}
	if len(page.Items) != 2 || page.Items[0].ID != "first" || page.Items[1].ID != "second" {
		t.Fatalf("unexpected tracks: %+v", page.Items)
	}
	image := page.Items[0].Album.Images[0]
	if image.Height != nil || image.Width != nil {
		t.Errorf("expected null image dimensions, got %+v", image)
	}
}

func TestSpotifyAPIDataEmptyResults(t *testing.T) {
	var data SpotifyAPIData
	if err := json.Unmarshal([]byte(`{"tracks":{"total":0,"items":[]}}`), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Tracks.Items == nil || len(data.Tracks.Items) != 0 {
		t.Fatalf("expected empty items array, got %+v", data.Tracks.Items)
	}
}

func TestStreamingSpotifyDataRejectsInvalidAPIData(t *testing.T) {
	for _, input := range []string{
		`{"streamingService":"spotify","trackUrl":"","trackId":"","apiData":{"tracks":`,
		`{"streamingService":"spotify","trackUrl":"","trackId":"","apiData":{"tracks":{"limit":"five"}}}`,
		`{"streamingService":"spotify","trackUrl":"","trackId":"","apiData":{"tracks":{"items":[{"duration_ms":"long"}]}}}`,
	} {
		t.Run(input, func(t *testing.T) {
			var data StreamingSpotifyData
			if err := json.Unmarshal([]byte(input), &data); err == nil {
				t.Fatal("expected an error for invalid Spotify API data")
			}
		})
	}
}
