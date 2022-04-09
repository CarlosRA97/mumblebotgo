package sourceProvider

type YoutubeDLSource struct {
	source string
	metadata *YoutubeDLSourceMetadata
}

type YoutubeDLSourceMetadata struct {
	ID                 string             `json:"id"`
	Title              string             `json:"title"`
	Duration           float64            `json:"duration"`
	IsLive             bool        		  `json:"is_live"`
	StartTime          interface{}        `json:"start_time"`
	EndTime            interface{}        `json:"end_time"`
	Series             interface{}        `json:"series"`
	Artist             interface{}        `json:"artist"`
	Album              interface{}        `json:"album"`
	ReleaseYear        interface{}        `json:"release_year"`
	Thumbnail          string             `json:"thumbnail"`
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
