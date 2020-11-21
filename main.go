package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

const (
	TempAudio = "downloaded.ogg"

	Play = "play"
	Pause = "pause"
	Stop = "stop"
	Search = "search"
	Volume = "v"
)

var (
	commandFunc = make(map[string]func(*gumble.TextMessageEvent))
	availableCommands = []string{Play, Pause, Stop, Search, Volume}
	tempAudioPath = path.Join(os.TempDir(), TempAudio)
)

func main() {
	os.Remove(tempAudioPath)
	var stream *gumbleffmpeg.Stream
	var offset time.Duration

	reCommand, err := regexp.Compile(fmt.Sprintf("(?:!)(%s)", strings.Join(availableCommands, "|")))
	if check(err) { return }
	reHref, err := regexp.Compile("(?:\")(https:|http:|www\\.)\\S*(?:\")")
	if check(err) { return }

	playSource := func (client *gumble.Client, source string) {
		if stream != nil {
			stream.Stop()
			os.Remove(tempAudioPath)
		}
		stream = gumbleffmpeg.New(client, gumbleffmpeg.SourceFile(tempAudioPath))
		stream.Volume = 0.05
		if err := stream.Play(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Playing %s\n", source)
		}
	}

	submatchExtract := func (match [][]string) (string, error) {
		err := errors.New("No match")
		fmt.Println(match)
		if len(match) > 0 && len(match[0]) > 1 {
			return match[0][1], nil
		}
		return "", err
	}

	commandFunc[Volume] = func(e *gumble.TextMessageEvent) {
		re, err := regexp.Compile(fmt.Sprintf("(?:!%s) *\\b(0|[1-9][0-9]?|100)\\b", Volume))
		check(err)
		number, err := submatchExtract(re.FindAllStringSubmatch(e.Message, -1))
		if stream != nil && err == nil {
			num, _ := strconv.ParseFloat(number, 32)
			stream.Volume = float32(num/100)
		}
		if stream != nil {
			sendMessage(e, fmt.Sprintf("Volume: %v%%\n", stream.Volume * 100))
		} else {
			sendMessage(e, "Nothing playing")
		}
	}

	commandFunc[Search] = func(e *gumble.TextMessageEvent) {
		re, err := regexp.Compile(fmt.Sprintf("(?:!%s) *(.*)", Search))
		check(err)
		searchMatch := re.FindAllStringSubmatch(e.Message, -1)
		search := try(submatchExtract(searchMatch))
		executeCommand(fmt.Sprintf("ytsearch1:%s", search))
		playSource(e.Client, search)
	}

	commandFunc[Play] = func(e *gumble.TextMessageEvent) {
		streamStateInitial := stream != nil && (
			stream.State() == 0	||
			stream.State() == gumbleffmpeg.StateInitial )
		streamStatePaused := stream != nil && 
			stream.State() == gumbleffmpeg.StatePaused

		if stream == nil {
			matchesLink := reHref.FindAllStringSubmatch(e.Message, -1)
			link, err := submatchExtract(matchesLink)
			fmt.Println(link)
			if !check(err) { 
				executeCommand(link)
			} else {
				return
			}
			playSource(e.Client, link)
			return
		}

		if streamStateInitial || streamStatePaused {
			stream.Offset = offset
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing\n")
			}
			return
		}
	}

	commandFunc[Pause] = func(e *gumble.TextMessageEvent) {
		streamStatePlaying := stream != nil &&
					stream.State() == gumbleffmpeg.StatePlaying
		fmt.Printf("Pause requirements: %v\n", streamStatePlaying)
		if streamStatePlaying {
			fmt.Println(e.Message)
			if err := stream.Pause(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				offset = stream.Offset
				fmt.Printf("Pausing\n")
			}
			return
		}
	}

	commandFunc[Stop] = func(e *gumble.TextMessageEvent) {
		streamStatePaused := stream != nil && 
					stream.State() == gumbleffmpeg.StatePaused
		streamStatePlaying := stream != nil &&
					stream.State() == gumbleffmpeg.StatePlaying
		fmt.Printf("Stop requirements: %v\n", streamStatePlaying || streamStatePaused)
		if streamStatePlaying || streamStatePaused {
			if err := stream.Stop(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Stopped\n")
				stream = nil
				if err := os.Remove("downloaded.ogg"); err != nil {
					fmt.Println("No se pudo borrar el archivo downloaded.ogg")
				}
			}
			return
		}
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] [audio files...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	gumbleutil.Main(gumbleutil.AutoBitrate, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			fmt.Println("Connected to the server")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}

			matchesCommand := reCommand.FindAllStringSubmatch(e.Message, -1)
			fmt.Println(matchesCommand)		
		
			command, err := submatchExtract(matchesCommand)
			// fmt.Println(command)
			if check(err) { return }
			commandFunc[command](e)
		},
	})
}

func try(msg string, err error) string {
	if check(err) { return "" }
	return msg
}

func check(err error) bool {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return true
	}
	return false
}

func executeCommand(param string) {
	cmd := exec.Command("youtube-dl","--extract-audio","--audio-format","vorbis", "--output", tempAudioPath,  param)
	// "ytsearch1: " + e.Message[1:] + ""
	if err := cmd.Run(); err == nil {
		fmt.Printf("Descargado %s\n", param)
	} else {
		fmt.Printf("%s\n", err)
	}
}

func sendMessage(e *gumble.TextMessageEvent, msg string) {
	e.Client.Send(&gumble.TextMessage{Message: msg, Channels: e.Channels, Sender: e.Sender, Users: e.Users})
}