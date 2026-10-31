package streaming

// StreamingService identifies a streaming provider and also permits custom names.
type StreamingService string

// Service names match the Python sister project's StreamingService enum.
const (
	Spotify           StreamingService = "spotify"
	TencentMusic      StreamingService = "tencent-music"
	YouTubeMusic      StreamingService = "youtube-music"
	IHeartRadio       StreamingService = "iheartradio"
	NetEaseCloudMusic StreamingService = "netease-cloud-music"
	AppleMusic        StreamingService = "apple-music"
	JioSaavn          StreamingService = "jiosaavn"
	AmazonMusic       StreamingService = "amazon-music"
	SoundCloud        StreamingService = "soundcloud"
	Anghami           StreamingService = "anghami"
	Gaana             StreamingService = "gaana"
	Pandora           StreamingService = "pandora"
	Boomplay          StreamingService = "boomplay"
	Deezer            StreamingService = "deezer"
	Mdundo            StreamingService = "mdundo"
	Audiomack         StreamingService = "audiomack"
	Melon             StreamingService = "melon"
	KakaoMusic        StreamingService = "kakao-music"
	LineMusic         StreamingService = "line-music"
	KKBox             StreamingService = "kkbox"
	Tidal             StreamingService = "tidal"
	YandexMusic       StreamingService = "yandex-music"
	Joox              StreamingService = "joox"
	Napster           StreamingService = "napster"
	Rhapsody          StreamingService = "rhapsody"
)
