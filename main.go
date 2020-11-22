package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
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
	tempAudio = "downloaded.ogg"

	playSong = "play"
	playPause = "p"
	stop = "stop"
	search = "search"
	queue = "q"
	volume = "v"
	elapsed = "e"

	wordsAfterCommand = "(\\w+[\\w| ]*)"
	numberNotGreaterThan100 = "\\b(0|[1-9][0-9]?|100)\\b"
	urlWithinDoubleQuotes = ".*(?:\")([https://|http://|www\\.]\\S*)(?:\")"
)

var (
	commandFunc = make(map[string]func(*gumble.TextMessageEvent))
	availableCommands = []string{playSong, playPause, stop, search, volume, elapsed, queue}
	tempAudioPath = path.Join(os.TempDir(), tempAudio)
	matchAvailableCommands = fmt.Sprintf("(%s)", strings.Join(availableCommands, "|"))
	queueSongs = make([]string, 0, 10)
	client *gumble.Client
)

func main() {
	var stream *gumbleffmpeg.Stream
	var offset time.Duration
	
	playSource := func (client *gumble.Client, source string) {
		if stream != nil {
			stream.Stop()
		}
		sourceAudio := gumbleffmpeg.SourceExec("youtube-dl", "-f", "bestaudio", "--rm-cache-dir", "-q", "-o", "-", source)
		stream = gumbleffmpeg.New(client, sourceAudio)
		stream.Volume = 0.05
		if err := stream.Play(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Playing %s\n", source)
		}
	}

	go func() {
		for {
			if stream != nil {
				switch stream.State() {
					case gumbleffmpeg.StatePlaying: {
						stream.Wait()
						fmt.Println("He terminado la cancion")
						if len(queueSongs) > 0 {
							fmt.Printf("Siguente cancion %s\n", queueSongs[0])
							playSource(client, queueSongs[0])
							queueSongs = queueSongs[1:]
						} else {
							fmt.Println("No hay mas canciones en la cola")
						}
					}; break
				}
			}
			time.Sleep(time.Second * 1)
		}
	}()

	submatchExtract := func (re *regexp.Regexp, message string) (string, error) {
		err := errors.New("No match")
		fmt.Println(message)
		match := re.FindAllStringSubmatch(message, -1)
		fmt.Println(match)
		if len(match) > 0 && len(match[0]) > 1 {
			return match[0][1], nil
		}
		return "", err
	}

	commandFunc[queue] = func(e *gumble.TextMessageEvent) {
		reQueueHref :=  regexAfterCommand(queue, urlWithinDoubleQuotes)
		reQueueSearch := regexAfterCommand(queue, wordsAfterCommand)
		if link, err := submatchExtract(reQueueHref, e.Message); !check(err) {
			queueSongs = append(queueSongs, link)
		}
		if search, err := submatchExtract(reQueueSearch, e.Message); !check(err) {
			queueSongs = append(queueSongs, fmt.Sprintf("ytsearch1:%s", search))
		}
		sendMessage(e, strings.Join(queueSongs, " -> "))
	}

	commandFunc[elapsed] = func(e *gumble.TextMessageEvent) {
		if stream != nil {
			sendMessage(e, stream.Elapsed().String())
		}
	}

	commandFunc[volume] = func(e *gumble.TextMessageEvent) {
		
		re := regexAfterCommand(volume, numberNotGreaterThan100)
		number, err := submatchExtract(re, e.Message)
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

	commandFunc[search] = func(e *gumble.TextMessageEvent) {
		re := regexAfterCommand(search, wordsAfterCommand)
		if search, err := submatchExtract(re, e.Message); !check(err) {
			playSource(e.Client, fmt.Sprintf("ytsearch1:%s", search))
		}
	}

	commandFunc[playSong] = func(e *gumble.TextMessageEvent) {
		rePlaySongHref := regexAfterCommand(playSong, urlWithinDoubleQuotes)
		if link, err := submatchExtract(rePlaySongHref, e.Message); !check(err) { 
			playSource(e.Client, link)
		}
	}

	commandFunc[playPause] = func(e *gumble.TextMessageEvent) {
		streamStatePlaying := stream != nil &&
					stream.State() == gumbleffmpeg.StatePlaying
		streamStatePaused := stream != nil && 
					stream.State() == gumbleffmpeg.StatePaused

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

		if streamStatePaused {
			stream.Offset = offset
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing\n")
			}
			return
		}
	}

	commandFunc[stop] = func(e *gumble.TextMessageEvent) {
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
			client = e.Client
			fmt.Println("Connected to the server")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			reCommand := regexAfterCommand(matchAvailableCommands, "")
			if e.Sender == nil {
				return
			}
		
			command, err := submatchExtract(reCommand, e.Message)
			if check(err) { return }
			commandFunc[command](e)
		},
	})
}

func check(err error) bool {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return true
	}
	return false
}

func regexAfterCommand(command string, expr string) *regexp.Regexp {
	return regexAfterCommandWithSpecialCharacter("", command, expr)
}

func regexAfterCommandWithSpecialCharacter(specialCharacter string, command string, expr string) *regexp.Regexp {
	if specialCharacter == "" { specialCharacter = "!" }
	return regexp.MustCompile(fmt.Sprintf("(?:%s%s) *%s", specialCharacter, command, expr))
}

func sendMessage(e *gumble.TextMessageEvent, msg string) {
	client.Send(&gumble.TextMessage{Message: msg, Channels: e.Channels, Sender: e.Sender, Users: e.Users})
}