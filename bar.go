package main

import (
	"strings"

	"layeh.com/gumble/gumbleffmpeg"
)

type ProgressBar struct {
	line string
	dot string
	stream *gumbleffmpeg.Stream
	streamDuration float64
	barLength int
}

func NewBar(stream *gumbleffmpeg.Stream, streamDuration float64) *ProgressBar {
	return NewBarWithLength(stream, streamDuration, 50)
}
func NewBarWithLength(stream *gumbleffmpeg.Stream, streamDuration float64, length int) *ProgressBar {
	return NewBarWithStyle(stream, streamDuration, length, "⎯", "○")

}
func NewBarWithStyle(stream *gumbleffmpeg.Stream, streamDuration float64, length int, line string, dot string) *ProgressBar {
	return &ProgressBar{line, dot, stream, streamDuration, length}
}

func (b *ProgressBar) generate() string {
	if b.stream == nil {
		return ""
	}
	normalizedValue := b.stream.Elapsed().Seconds() / b.streamDuration
	index := int(normalizedValue * float64(b.barLength))
	return strings.Join([]string{strings.Repeat(b.line, index), strings.Repeat(b.line, b.barLength - index - 1)}, b.dot)
}