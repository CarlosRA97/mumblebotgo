package player

import (
	"errors"
	"fmt"
	"log"
	"mumblebot/sourceProvider"
	"strings"
	"sync"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
)

func NewPlayer() *Player {
	return &Player{
		SourceProvider: &sourceProvider.YoutubeDLSource{},
		queue: make([]string, 0, 10),
		client: nil,
		stream: nil,
		volume: 0.05,
		metadataAvailable: false,
		progressBar: nil,
		offset: 0,
		wg: sync.WaitGroup{},
	}
}

func (p *Player) setSourceProvider(provider sourceProvider.ISourceProvider) {
	p.SourceProvider = provider
}

func (p *Player) IsMetadataAvailable() bool {
	return p.metadataAvailable
}

func (p *Player) Elapsed() string {
	if p.HasStream() {
		return p.stream.Elapsed().String()
	}
	return ""
}

func (p *Player) PlayOrQueue(source string, callback func(status string)) {
	if p.IsStopped() {
		p.Play(source)
		if len(strings.Split(source, "1:")) > 1 {
			callback(fmt.Sprintf("Playing %s\n", strings.Split(source, "1:")[1]))
		} else {
			callback(fmt.Sprintf("Playing %s\n", source))
		}
	} else {
		p.Enqueue(source)
		callback(strings.Join(p.queue, " -> "))
	}
}

func (p *Player) metadataCheckIfAvailable() {
	p.wg.Add(1)
	if err := p.SourceProvider.Metadata(); !check(err) {
		p.metadataAvailable = true
		p.wg.Done()
	}
}

func (p *Player) Play(source string) {
	if p.HasStream() {
		p.stream.Stop()
	}
	
	p.SourceProvider.SetSource(source)

	go p.metadataCheckIfAvailable()

	p.wg.Add(1)
	p.stream = gumbleffmpeg.New(p.client, p.SourceProvider.Source())
	p.stream.Volume = p.volume
	p.wg.Done()
	
	go func () {
		log.Println("gofunc progressbar")
		p.wg.Wait()
		p.progressBar = NewBar(p.stream, p.SourceProvider.MetadataDuration())
		p.metadataAvailable = true
	}()
	
	if err := p.stream.Play(); err != nil {
		log.Printf("%s\n", err)
		return
	}
	
	log.Printf("Playing %s\n", source)
}

func (p *Player) Stop(callback func (status string)) {
	p.queue = make([]string, 0, 10)
	if p.HasStream() {
		p.stream.Stop()
		p.stream = nil
		p.metadataAvailable = false
		if callback != nil {
			callback("Stopped")
		}
	}
}

func (p *Player) Skip() {
	if p.IsPlayingOrPaused() {
		if err := p.stream.Stop(); err != nil {
			log.Printf("%s\n", err)
		} else {
			log.Printf("Skipped\n")
			p.stream = nil
			p.metadataAvailable = false
		}
	}
}

func (p *Player) Queue() []string {
	return p.queue
}

func (p *Player) Enqueue(source string) {
	p.queue = append(p.queue, source)
}

func (p *Player) Dequeue() (value string, err error) {
	if len(p.queue) > 0 {
		value = p.queue[0]
		p.queue = p.queue[1:]
		return
	}
	return "", errors.New("No more items")
}

func (p *Player) PlayPause(callback func (status string)) {
	streamStatePlaying := p.HasStream() &&
				p.stream.State() == gumbleffmpeg.StatePlaying
	streamStatePaused := p.HasStream() && 
				p.stream.State() == gumbleffmpeg.StatePaused

	if streamStatePlaying {
		if err := p.stream.Pause(); err != nil {
			log.Printf("%s\n", err)
		} else {
			p.offset = p.stream.Offset
			log.Printf("Pausing\n")
			callback("Pause")	
		}
	}

	if streamStatePaused {
		p.stream.Offset = p.offset
		if err := p.stream.Play(); err != nil {
			log.Printf("%s\n", err)
		} else {
			log.Printf("Playing\n")
			callback("Playing")
		}
	}
}

func (p *Player) HasStream() bool {
	return p.stream != nil
}

func (p *Player) IsStopped() bool {
	return !p.HasStream() || (p.HasStream() && p.stream.State() == gumbleffmpeg.StateStopped)
}

func (p *Player) IsPlayingOrPaused() bool {
	streamStatePaused := p.HasStream() && 
					p.stream.State() == gumbleffmpeg.StatePaused
	streamStatePlaying := p.HasStream() &&
					p.stream.State() == gumbleffmpeg.StatePlaying
	return streamStatePlaying || streamStatePaused
}

func (p *Player) Progress() string {
	return p.progressBar.generate()
}

func (p *Player) SetClient(client *gumble.Client) {
	p.client = client
}

func (p *Player) SetVolume(volume float32) {
	if p.HasStream() {
		p.stream.Volume = volume
	}
	p.volume = volume
}

func (p *Player) NormalizedVolume() int {
	if p.HasStream() { 
		return int(p.stream.Volume * 100)
	}
	return int(p.volume * 100)
}

func check(err error) bool {
	if err == nil {
		return false
	}
	log.Printf("%s\n", err)
	return true
}