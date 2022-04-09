package player

import (
	"mumblebot/sourceProvider"
	"sync"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
)

type Player struct {
	SourceProvider sourceProvider.ISourceProvider
	queue []string
	client *gumble.Client
	stream *gumbleffmpeg.Stream
	volume float32
	metadataAvailable bool
	progressBar *ProgressBar
	offset time.Duration
	wg sync.WaitGroup
}