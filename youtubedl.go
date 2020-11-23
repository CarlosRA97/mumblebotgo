package main

import "encoding/json"

type SourceMetadata struct {
	ID                 string             `json:"id"`
	Uploader           string             `json:"uploader"`
	UploaderID         string             `json:"uploader_id"`
	UploaderURL        string             `json:"uploader_url"`
	ChannelID          string             `json:"channel_id"`
	ChannelURL         string             `json:"channel_url"`
	UploadDate         string             `json:"upload_date"`
	License            interface{}        `json:"license"`
	Creator            interface{}        `json:"creator"`
	Title              string             `json:"title"`
	AltTitle           interface{}        `json:"alt_title"`
	Thumbnails         []Thumbnails       `json:"thumbnails"`
	Description        string             `json:"description"`
	Categories         []string           `json:"categories"`
	Tags               []string           `json:"tags"`
	Subtitles          Subtitles          `json:"subtitles"`
	AutomaticCaptions  AutomaticCaptions  `json:"automatic_captions"`
	Duration           float64            `json:"duration"`
	AgeLimit           int                `json:"age_limit"`
	Annotations        interface{}        `json:"annotations"`
	Chapters           interface{}        `json:"chapters"`
	WebpageURL         string             `json:"webpage_url"`
	ViewCount          int                `json:"view_count"`
	LikeCount          interface{}        `json:"like_count"`
	DislikeCount       interface{}        `json:"dislike_count"`
	AverageRating      float64            `json:"average_rating"`
	Formats            []Formats          `json:"formats"`
	IsLive             interface{}        `json:"is_live"`
	StartTime          interface{}        `json:"start_time"`
	EndTime            interface{}        `json:"end_time"`
	Series             interface{}        `json:"series"`
	SeasonNumber       interface{}        `json:"season_number"`
	EpisodeNumber      interface{}        `json:"episode_number"`
	Track              interface{}        `json:"track"`
	Artist             interface{}        `json:"artist"`
	Album              interface{}        `json:"album"`
	ReleaseDate        interface{}        `json:"release_date"`
	ReleaseYear        interface{}        `json:"release_year"`
	Extractor          string             `json:"extractor"`
	WebpageURLBasename string             `json:"webpage_url_basename"`
	ExtractorKey       string             `json:"extractor_key"`
	NEntries           int                `json:"n_entries"`
	Playlist           string             `json:"playlist"`
	PlaylistID         string             `json:"playlist_id"`
	PlaylistTitle      interface{}        `json:"playlist_title"`
	PlaylistUploader   interface{}        `json:"playlist_uploader"`
	PlaylistUploaderID interface{}        `json:"playlist_uploader_id"`
	PlaylistIndex      int                `json:"playlist_index"`
	Thumbnail          string             `json:"thumbnail"`
	DisplayID          string             `json:"display_id"`
	RequestedSubtitles interface{}        `json:"requested_subtitles"`
	RequestedFormats   []RequestedFormats `json:"requested_formats"`
	Format             string             `json:"format"`
	FormatID           string             `json:"format_id"`
	Width              int                `json:"width"`
	Height             int                `json:"height"`
	Resolution         interface{}        `json:"resolution"`
	Fps                int                `json:"fps"`
	Vcodec             string             `json:"vcodec"`
	Vbr                interface{}        `json:"vbr"`
	StretchedRatio     interface{}        `json:"stretched_ratio"`
	Acodec             string             `json:"acodec"`
	Abr                int                `json:"abr"`
	Ext                string             `json:"ext"`
	Fulltitle          string             `json:"fulltitle"`
	Filename           string             `json:"_filename"`
}
type Thumbnails struct {
	URL        string `json:"url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Resolution string `json:"resolution"`
	ID         string `json:"id"`
}
type Subtitles struct {
}
type AutomaticCaptions struct {
}
type DownloaderOptions struct {
	HTTPChunkSize int `json:"http_chunk_size"`
}
type HTTPHeaders struct {
	UserAgent      string `json:"User-Agent"`
	AcceptCharset  string `json:"Accept-Charset"`
	Accept         string `json:"Accept"`
	AcceptEncoding string `json:"Accept-Encoding"`
	AcceptLanguage string `json:"Accept-Language"`
}
type Formats struct {
	FormatID          string            `json:"format_id"`
	URL               string            `json:"url"`
	PlayerURL         string            `json:"player_url"`
	Ext               string            `json:"ext"`
	FormatNote        string            `json:"format_note"`
	Acodec            string            `json:"acodec"`
	Abr               int               `json:"abr,omitempty"`
	Asr               int               `json:"asr"`
	Filesize          int               `json:"filesize"`
	Fps               interface{}       `json:"fps"`
	Height            interface{}       `json:"height"`
	Tbr               float64           `json:"tbr"`
	Width             interface{}       `json:"width"`
	Vcodec            string            `json:"vcodec"`
	DownloaderOptions DownloaderOptions `json:"downloader_options,omitempty"`
	Format            string            `json:"format"`
	Protocol          string            `json:"protocol"`
	HTTPHeaders       HTTPHeaders       `json:"http_headers"`
	Container         string            `json:"container,omitempty"`
}
type RequestedFormats struct {
	FormatID          string            `json:"format_id"`
	URL               string            `json:"url"`
	PlayerURL         string            `json:"player_url"`
	Ext               string            `json:"ext"`
	Height            int               `json:"height"`
	FormatNote        string            `json:"format_note"`
	Vcodec            string            `json:"vcodec"`
	Asr               interface{}       `json:"asr"`
	Filesize          int               `json:"filesize"`
	Fps               int               `json:"fps"`
	Tbr               float64           `json:"tbr"`
	Width             int               `json:"width"`
	Acodec            string            `json:"acodec"`
	DownloaderOptions DownloaderOptions `json:"downloader_options"`
	Format            string            `json:"format"`
	Protocol          string            `json:"protocol"`
	HTTPHeaders       HTTPHeaders       `json:"http_headers"`
	Abr               int               `json:"abr,omitempty"`
}

func parseSourceMetadata(metadata string) (*SourceMetadata, error) {
	sourceMetadata := &SourceMetadata{}
	if err := json.Unmarshal([]byte(metadata), sourceMetadata); err != nil {
		return nil, err
	}
	return sourceMetadata, nil
}