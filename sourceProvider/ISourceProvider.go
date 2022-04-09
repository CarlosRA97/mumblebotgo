package sourceProvider

import "layeh.com/gumble/gumbleffmpeg"

type ISourceProvider interface {
	SetSource(string)
	Source() gumbleffmpeg.Source
	Metadata() error
	MetadataTitle() string
	MetadataDuration() float64
	MetadataIsLive() bool
}