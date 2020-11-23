package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	defautlVolume = 0.05

	nowPlaying = "np"
	playSong = "play"
	playPause = "p"
	skip = "skip"
	stop = "stop"
	search = "search"
	queue = "q"
	volume = "v"
	elapsed = "e"
	help = "h"

	wordsAfterCommand = "(\\w+[\\w| ]*)"
	numberNotGreaterThan100 = "\\b(0|[1-9][0-9]?|100)\\b"
	urlWithinDoubleQuotes = ".*(?:\")([https://|http://|www\\.]\\S*)(?:\")"
)

var (
	commandFunc = make(map[string]func(*gumble.TextMessageEvent))
	availableCommands = []string{help, playSong, playPause, nowPlaying, skip, stop, search, volume, elapsed, queue}
	matchAvailableCommands = fmt.Sprintf("(%s)\\b", strings.Join(availableCommands, "|"))
	queueSongs = make([]string, 0, 10)
)

func main() {
	var stream *gumbleffmpeg.Stream
	var offset time.Duration

	reCommand := regexAfterCommand(matchAvailableCommands, "")

	reQueueHref :=  regexAfterCommand(queue, urlWithinDoubleQuotes)
	reQueueSearch := regexAfterCommand(queue, wordsAfterCommand)
	reVolumeWithIn100 := regexAfterCommand(volume, numberNotGreaterThan100)
	rePlaySongSearch := regexAfterCommand(playSong, wordsAfterCommand)
	rePlaySongHref := regexAfterCommand(playSong, urlWithinDoubleQuotes)
	
	playSource := func (client *gumble.Client, source string) {
		if stream != nil {
			stream.Stop()
		}
		sourceAudio := gumbleffmpeg.SourceExec("youtube-dl", "-f", "bestaudio", "--rm-cache-dir", "-q", "-o", "-", source)
		stream = gumbleffmpeg.New(client, sourceAudio)
		stream.Volume = defautlVolume
		if err := stream.Play(); err != nil {
			fmt.Printf("%s\n", err)
		} else {
			fmt.Printf("Playing %s\n", source)
		}
	}

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

	commandFunc[help] = func(e *gumble.TextMessageEvent) {
		sendMessage(e, fmt.Sprintf("Comandos: %s\n", strings.Join(availableCommands[1:], ", ")))
	}

	commandFunc[nowPlaying] = func(e *gumble.TextMessageEvent) {

		stream.Elapsed().Seconds()
		sendMessage(e, "<h2>Title</h2> (==#==========)")
	}

	commandFunc[queue] = func(e *gumble.TextMessageEvent) {
		if link, err := submatchExtract(reQueueHref, e.Message); !check(err) {
			queueSongs = append(queueSongs, link)
		}
		if search, err := submatchExtract(reQueueSearch, e.Message); !check(err) {
			queueSongs = append(queueSongs, fmt.Sprintf("ytsearch1:%s", search))
		}
		if len(queueSongs) == 0 {
			sendMessage(e, "No queued songs")
		}
		sendMessage(e, strings.Join(queueSongs, " -> "))
	}

	commandFunc[elapsed] = func(e *gumble.TextMessageEvent) {
		if stream != nil {
			sendMessage(e, stream.Elapsed().String())
		}
	}

	commandFunc[volume] = func(e *gumble.TextMessageEvent) {
		
		if number, err := submatchExtract(reVolumeWithIn100, e.Message); stream != nil && err == nil {
			num, _ := strconv.ParseFloat(number, 32)
			stream.Volume = float32(num/100)
		}

		volumeStreamNormalized := func () float32 {
			if stream != nil { 
				return stream.Volume * 100
			}
			return defautlVolume * 100
		}

		sendMessage(e, fmt.Sprintf("Volume: %v%%\n", volumeStreamNormalized()))
	}

	commandFunc[search] = func(e *gumble.TextMessageEvent) {
		sendMessage(e, "Aqui saldran las busquedas de youtube")
	}
	
	commandFunc[playSong] = func(e *gumble.TextMessageEvent) {
		playOrQueue := func (source string) {
			if stream == nil || (stream != nil && stream.State() == gumbleffmpeg.StateStopped) {
				playSource(e.Client, source)
				if len(strings.Split(source, "1:")) > 1 {
					sendMessage(e, fmt.Sprintf("Playing %s\n", strings.Split(source, "1:")[1]))
				} else {
					sendMessage(e, fmt.Sprintf("Playing %s\n", source))
				}
			} else {
				queueSongs = append(queueSongs, source)
				sendMessage(e, strings.Join(queueSongs, " -> "))
			}
		}

		if link, err := submatchExtract(rePlaySongHref, e.Message); !check(err) { 
			playOrQueue(link)
		}
		if search, err := submatchExtract(rePlaySongSearch, e.Message); !check(err) {
			playOrQueue(fmt.Sprintf("ytsearch1:%s", search))
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
				sendMessage(e, "Pause")
			}
			return
		}

		if streamStatePaused {
			stream.Offset = offset
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing\n")
				sendMessage(e, "Playing")
			}
			return
		}
	}

	commandFunc[stop] = func(e *gumble.TextMessageEvent) {
		queueSongs = make([]string, 0, 10)
		if stream != nil {
			stream.Stop()
			stream = nil
			sendMessage(e, "Stopped")
		}
	}

	commandFunc[skip] = func(e *gumble.TextMessageEvent) {
		streamStatePaused := stream != nil && 
					stream.State() == gumbleffmpeg.StatePaused
		streamStatePlaying := stream != nil &&
					stream.State() == gumbleffmpeg.StatePlaying
		fmt.Printf("Skip requirements: %v\n", streamStatePlaying || streamStatePaused)
		if streamStatePlaying || streamStatePaused {
			if err := stream.Stop(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Skipped\n")
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
			go func(client *gumble.Client) {
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
			}(e.Client)
			fmt.Println("Connected to the server")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
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

func executeCommand(param string)  {
	cmd := exec.Command("youtube-dl","-j", param)
	if err := cmd.Run(); !check(err) {
		
	}
}

func regexAfterCommand(command string, expr string) *regexp.Regexp {
	return regexAfterCommandWithSpecialCharacter("", command, expr)
}

func regexAfterCommandWithSpecialCharacter(specialCharacter string, command string, expr string) *regexp.Regexp {
	if specialCharacter == "" { specialCharacter = "!" }
	return regexp.MustCompile(fmt.Sprintf("(?:%s%s) *%s", specialCharacter, command, expr))
}

func sendMessage(e *gumble.TextMessageEvent, msg string) {
	e.Client.Send(&gumble.TextMessage{Message: msg, Channels: e.Channels, Sender: e.Sender, Users: e.Users})
}