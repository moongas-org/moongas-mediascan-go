package streaming

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStreamingServiceSerialization(t *testing.T) {
	services := []struct {
		service StreamingService
		value   string
	}{
		{Spotify, "spotify"},
		{TencentMusic, "tencent-music"},
		{YouTubeMusic, "youtube-music"},
		{IHeartRadio, "iheartradio"},
		{NetEaseCloudMusic, "netease-cloud-music"},
		{AppleMusic, "apple-music"},
		{JioSaavn, "jiosaavn"},
		{AmazonMusic, "amazon-music"},
		{SoundCloud, "soundcloud"},
		{Anghami, "anghami"},
		{Gaana, "gaana"},
		{Pandora, "pandora"},
		{Boomplay, "boomplay"},
		{Deezer, "deezer"},
		{Mdundo, "mdundo"},
		{Audiomack, "audiomack"},
		{Melon, "melon"},
		{KakaoMusic, "kakao-music"},
		{LineMusic, "line-music"},
		{KKBox, "kkbox"},
		{Tidal, "tidal"},
		{YandexMusic, "yandex-music"},
		{Joox, "joox"},
		{Napster, "napster"},
		{Rhapsody, "rhapsody"},
		{StreamingService("custom-service"), "custom-service"},
	}
	for _, tt := range services {
		t.Run(tt.value, func(t *testing.T) {
			if string(tt.service) != tt.value {
				t.Fatalf("service = %q, want %q", tt.service, tt.value)
			}
			for _, codec := range []struct {
				name      string
				marshal   func(any) ([]byte, error)
				unmarshal func([]byte, any) error
			}{
				{"JSON", json.Marshal, json.Unmarshal},
				{"YAML", yaml.Marshal, yaml.Unmarshal},
			} {
				t.Run(codec.name, func(t *testing.T) {
					data := StreamingData[any]{StreamingService: tt.service}
					out, err := codec.marshal(data)
					if err != nil {
						t.Fatalf("marshal: %v", err)
					}
					var fields map[string]any
					if err := codec.unmarshal(out, &fields); err != nil {
						t.Fatalf("unmarshal fields: %v", err)
					}
					if fields["streamingService"] != tt.value {
						t.Errorf("unexpected streamingService: %#v", fields["streamingService"])
					}
					var decoded StreamingData[any]
					if err := codec.unmarshal(out, &decoded); err != nil {
						t.Fatalf("unmarshal: %v", err)
					}
					if decoded.StreamingService != tt.service {
						t.Errorf("service = %q, want %q", decoded.StreamingService, tt.service)
					}
				})
			}
		})
	}
}
