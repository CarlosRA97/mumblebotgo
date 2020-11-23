package main

import (
	"errors"
	"flag"
	"fmt"
	"mumblebot/sourceProvider"
	"os"
	"regexp"
	"strconv"
	"strings"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

const (
	nowPlaying = "np"
	playSong = "play"
	playPause = "p"
	skip = "skip"
	stop = "stop"
	search = "search"
	queue = "q"
	volume = "v"
	help = "h"

	wordsAfterCommand = "(\\w+[\\w| ]*)"
	numberNotGreaterThan100 = "\\b(0|[1-9][0-9]?|100)\\b"
	urlWithinDoubleQuotes = ".*(?:\")([https://|http://|www\\.]\\S*)(?:\")"
)

func main() {	
	commands := NewCommands()
	player := NewPlayer()

	reQueueHref :=  regexAfterCommand(queue, urlWithinDoubleQuotes)
	reQueueSearch := regexAfterCommand(queue, wordsAfterCommand)
	reVolumeWithIn100 := regexAfterCommand(volume, numberNotGreaterThan100)
	rePlaySongSearch := regexAfterCommand(playSong, wordsAfterCommand)
	rePlaySongHref := regexAfterCommand(playSong, urlWithinDoubleQuotes)

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

	commands.add(nowPlaying, func(e *gumble.TextMessageEvent) {
		if !player.hasStream() {
			sendMessage(e, "Nothing playing")
			return
		}
		if player.currentlyPlayingSong == nil {
			sendMessage(e, "Searching info")
			return
		}

		sendMessage(e, fmt.Sprintf("<h2>%s</h2> ❨%s❩ %s", player.currentlyPlayingSong.(*sourceProvider.YoutubeDLSourceMetadata).Title, player.progress(), player.stream.Elapsed().String()))
	})

	commands.add(queue, func(e *gumble.TextMessageEvent) {
		if link, err := submatchExtract(reQueueHref, e.Message); !check(err) {
			player.enqueue(link)
		}
		if search, err := submatchExtract(reQueueSearch, e.Message); !check(err) {
			player.enqueue(fmt.Sprintf("ytsearch1:%s", search))
		}
		if len(player.queue) == 0 {
			sendMessage(e, "No queued songs")
		}
		sendMessage(e, strings.Join(player.queue, " -> "))
	})

	commands.add(volume, func(e *gumble.TextMessageEvent) {
		
		if number, err := submatchExtract(reVolumeWithIn100, e.Message); player.hasStream() && err == nil {
			num, _ := strconv.ParseFloat(number, 32)
			player.setVolume(float32(num/100))
		}

		sendMessage(e, fmt.Sprintf("Volume: %v%%\n", player.normalizedVolume()))
	})

	commands.add(search, func(e *gumble.TextMessageEvent) {
		sendMessage(e, "Aqui saldran las busquedas de youtube")
	})
	
	commands.add(playSong, func(e *gumble.TextMessageEvent) {
		send := func(status string) {
			sendMessage(e, status)
		}
		if link, err := submatchExtract(rePlaySongHref, e.Message); !check(err) { 
			player.playOrQueue(link, send)
		}
		if search, err := submatchExtract(rePlaySongSearch, e.Message); !check(err) {
			player.playOrQueue(fmt.Sprintf("ytsearch1:%s", search), send)
		}
	})

	commands.add(playPause, func(e *gumble.TextMessageEvent) {
		player.playPause(func(status string) {
			sendMessage(e, status)
		})
	})

	commands.add(stop, func(e *gumble.TextMessageEvent) {
		player.stop(func(status string) {
			sendMessage(e, status)
		})
	})

	commands.add(skip, func(e *gumble.TextMessageEvent) {
		player.skip()
	})

	commands.add(help, func(e *gumble.TextMessageEvent) {
		sendMessage(e, fmt.Sprintf("Comandos: %s<br>Mas informacion en <a href='https://rythmbot.co/features#list'>Rythmbot Command List</a>", strings.Join(commands.getAll(), ", ")))
	})

	matchAvailableCommands := fmt.Sprintf("(%s)\\b", strings.Join(commands.getAll(), "|"))
	reCommand := regexAfterCommand(matchAvailableCommands, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] [audio files...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	gumbleutil.Main(gumbleutil.AutoBitrate, gumbleutil.Listener{
		Connect: func(e *gumble.ConnectEvent) {
			player.setClient(e.Client)
			go player.queueHandler()
			fmt.Println("Connected to the server")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}
			command, err := submatchExtract(reCommand, e.Message)
			if check(err) { return }
			commands.execute(command, e)
		},
	})
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

func check(err error) bool {
	if err == nil {
		return false
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	return true
}