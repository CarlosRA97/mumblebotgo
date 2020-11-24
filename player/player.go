package player

import (
	"errors"
	"fmt"
	"mumblebot/sourceProvider"
	"os"
	"strings"
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

func (p *Player) Play(source string) {
	if p.HasStream() {
		p.stream.Stop()
	}
	
	p.SourceProvider.SetSource(source)

	go func() {
		p.wg.Add(1)
		if err := p.SourceProvider.Metadata(); !check(err) {
			p.metadataAvailable = true
			p.wg.Done()
		}
	}()

	p.wg.Add(1)
	p.stream = gumbleffmpeg.New(p.client, p.SourceProvider.Source())
	p.stream.Volume = p.volume
	p.wg.Done()
	
	go func () {
		fmt.Println("gofunc progressbar")
		p.wg.Wait()
		p.progressBar = NewBar(p.stream, p.SourceProvider.MetadataDuration())
		p.metadataAvailable = true
	}()
	
	if err := p.stream.Play(); err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	
	fmt.Printf("Playing %s\n", source)
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
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Skipped\n")
			p.stream = nil
			p.metadataAvailable = false
		}
	}
}

func (p *Player) Queue() []string {
	return p.queue
}

func (p *Player) QueueHandler() {
	for {
		if p.HasStream() {
			switch p.stream.State() {
				case gumbleffmpeg.StatePlaying: {
					p.stream.Wait()
					fmt.Println("He terminado la cancion")
					p.Skip()
					if source, err := p.Dequeue(); err == nil {
						fmt.Printf("Siguente cancion %s\n", source)
						p.Play(source)
					} else {
						p.Stop(nil)
						fmt.Println("No hay mas canciones en la cola")
					}
				}; break
			}
		}
		time.Sleep(time.Second * 1)
	}	
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
	fmt.Fprintf(os.Stderr, "%s\n", err)
	return true
}