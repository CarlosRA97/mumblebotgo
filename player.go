package main

import (
	"errors"
	"fmt"
	"mumblebot/sourceProvider"
	"strings"
	"sync"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
)

var wg = sync.WaitGroup{}

type Player struct {
	sourceProvider sourceProvider.ISourceProvider
	queue []string
	client *gumble.Client
	stream *gumbleffmpeg.Stream
	currentlyPlayingSong interface{}
	volume float32
	progressBar *ProgressBar
	offset time.Duration
}

func NewPlayer() *Player {
	return &Player{
		sourceProvider: &sourceProvider.YoutubeDLSource{},
		queue: make([]string, 0, 10),
		client: nil,
		stream: nil,
		currentlyPlayingSong: nil,
		volume: 0.05,
		progressBar: nil,
		offset: 0,
	}
}

func (p *Player) setSourceProvider(provider sourceProvider.ISourceProvider) {
	p.sourceProvider = provider
}

func (p *Player) playOrQueue(source string, callback func(status string)) {
	if p.isStopped() {
		p.play(source)
		if len(strings.Split(source, "1:")) > 1 {
			callback(fmt.Sprintf("Playing %s\n", strings.Split(source, "1:")[1]))
		} else {
			callback(fmt.Sprintf("Playing %s\n", source))
		}
	} else {
		p.enqueue(source)
		callback(strings.Join(p.queue, " -> "))
	}
}

func (p *Player) play(source string) {
	if p.hasStream() {
		p.stream.Stop()
	}
	
	p.sourceProvider.SetSource(source)

	go func() {
		wg.Add(1)
		if currentlyPlayingSongMetadata, err := p.sourceProvider.SourceMetadata(); !check(err) {
			p.currentlyPlayingSong = currentlyPlayingSongMetadata
		}
		wg.Done()
	}()

	wg.Add(1)
	p.stream = gumbleffmpeg.New(p.client, p.sourceProvider.Source())
	p.stream.Volume = p.volume
	wg.Done()
	
	go func () {
		wg.Wait()
		p.progressBar = NewBar(p.stream, p.currentlyPlayingSong.(*sourceProvider.YoutubeDLSourceMetadata).Duration)
	}()
	
	if err := p.stream.Play(); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	
	fmt.Printf("Playing %s\n", source)
}

func (p *Player) stop(callback func (status string)) {
	p.queue = make([]string, 0, 10)
	if p.hasStream() {
		p.stream.Stop()
		p.stream = nil
		p.currentlyPlayingSong = nil
		if callback != nil {
			callback("Stopped")
		}
	}
}

func (p *Player) skip() {
	if p.isPlayingOrPaused() {
		if err := p.stream.Stop(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Skipped\n")
			p.stream = nil
			p.currentlyPlayingSong = nil
		}
	}
}

func (p *Player) queueHandler() {
	for {
		if p.hasStream() {
			switch p.stream.State() {
				case gumbleffmpeg.StatePlaying: {
					p.stream.Wait()
					fmt.Println("He terminado la cancion")
					p.skip()
					if source, err := p.dequeue(); err == nil {
						fmt.Printf("Siguente cancion %s\n", source)
						p.play(source)
					} else {
						p.stop(nil)
						fmt.Println("No hay mas canciones en la cola")
					}
				}; break
			}
		}
		time.Sleep(time.Second * 1)
	}	
}

func (p *Player) enqueue(source string) {
	p.queue = append(p.queue, source)
}

func (p *Player) dequeue() (value string, err error) {
	if len(p.queue) > 0 {
		value = p.queue[0]
		p.queue = p.queue[1:]
		return
	}
	return "", errors.New("No more items")
}

func (p *Player) playPause(callback func (status string)) {
	streamStatePlaying := p.hasStream() &&
				p.stream.State() == gumbleffmpeg.StatePlaying
	streamStatePaused := p.hasStream() && 
				p.stream.State() == gumbleffmpeg.StatePaused

	if streamStatePlaying {
		if err := p.stream.Pause(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			p.offset = p.stream.Offset
			fmt.Printf("Pausing\n")
			callback("Pause")	
		}
	}

	if streamStatePaused {
		p.stream.Offset = p.offset
		if err := p.stream.Play(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Playing\n")
			callback("Playing")
		}
	}
}

func (p *Player) hasStream() bool {
	return p.stream != nil
}

func (p *Player) isStopped() bool {
	return !p.hasStream() || (p.hasStream() && p.stream.State() == gumbleffmpeg.StateStopped)
}

func (p *Player) isPlayingOrPaused() bool {
	streamStatePaused := p.hasStream() && 
					p.stream.State() == gumbleffmpeg.StatePaused
	streamStatePlaying := p.hasStream() &&
					p.stream.State() == gumbleffmpeg.StatePlaying
	return streamStatePlaying || streamStatePaused
}

func (p *Player) progress() string {
	return p.progressBar.generate()
}

func (p *Player) setClient(client *gumble.Client) {
	p.client = client
}

func (p *Player) setVolume(volume float32) {
	if p.hasStream() {
		p.stream.Volume = volume
	}
	p.volume = volume
}

func (p *Player) normalizedVolume() int {
	if p.hasStream() { 
		return int(p.stream.Volume * 100)
	}
	return int(p.volume * 100)
}