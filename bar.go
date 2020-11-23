package main

import (
	"math"
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
	normalizedValue := math.Abs(b.stream.Elapsed().Seconds()) / math.Abs(b.streamDuration)
	firstPart := int(math.Ceil(math.Abs(normalizedValue * float64(b.barLength))))
	secondPart := int(math.Abs(float64(b.barLength - firstPart - 1)))
	return strings.Join([]string{strings.Repeat(b.line, firstPart), strings.Repeat(b.line, secondPart)}, b.dot)
}