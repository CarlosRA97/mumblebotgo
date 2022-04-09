package sourceProvider

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os/exec"

	"layeh.com/gumble/gumbleffmpeg"
)

const ytdlBin = "yt-dlp"

func parseSourceMetadata(metadata string) (*YoutubeDLSourceMetadata, error) {
	sourceMetadata := &YoutubeDLSourceMetadata{}
	if err := json.Unmarshal([]byte(metadata), sourceMetadata); err != nil {
		return nil, err
	}
	return sourceMetadata, nil
}

func (s *YoutubeDLSource) SetSource(value string) {
	s.source = value
}

func (s *YoutubeDLSource) Source() gumbleffmpeg.Source {
	return gumbleffmpeg.SourceExec(ytdlBin, "-q", "-o", "-", s.source)
}

func (s *YoutubeDLSource) Metadata() (error) {
	cmd := exec.Command(ytdlBin, "-j", s.source)
	var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if check(err) {
		return err
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	log.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	if metadata, err := parseSourceMetadata(outStr); errStr == "" && !check(err) {
		s.metadata = metadata
		return nil
	}
	return errors.New(errStr)
}

func (s *YoutubeDLSource) MetadataTitle() string {
	if s.metadata != nil {
		return s.metadata.Title
	}
	return ""
}

func (s *YoutubeDLSource) MetadataDuration() float64 {
	if s.metadata != nil && !s.metadata.IsLive {
		return s.metadata.Duration
	}
	return 0
}

func (s *YoutubeDLSource) MetadataIsLive() bool {
	if s.metadata != nil {
		return s.metadata.IsLive
	}
	return false
}

func check(err error) bool {
	if err == nil {
		return false
	}
	log.Printf("[Error]:\n%s\n", err)
	return true
}

