package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"mumblebot/player"
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
	player := player.NewPlayer()

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
		if !player.HasStream() {
			sendMessage(e, "Nothing playing")
			return
		}
		if !player.IsMetadataAvailable() {
			sendMessage(e, "Searching info")
			return
		}

		if player.SourceProvider.MetadataIsLive() {
			sendMessage(e, fmt.Sprintf("<h2>%s</h2> <h3><strong>It's Live</strong></h3>", player.SourceProvider.MetadataTitle()))
			return
		}

		sendMessage(e, fmt.Sprintf("<h2>%s</h2> ❨%s❩ %s", player.SourceProvider.MetadataTitle(), player.Progress(), player.Elapsed()))
		
	})

	commands.add(queue, func(e *gumble.TextMessageEvent) {
		if link, err := submatchExtract(reQueueHref, e.Message); !check(err) {
			player.Enqueue(link)
		}
		if search, err := submatchExtract(reQueueSearch, e.Message); !check(err) {
			player.Enqueue(fmt.Sprintf("ytsearch1:%s", search))
		}
		if len(player.Queue()) == 0 {
			sendMessage(e, "No queued songs")
		}
		sendMessage(e, strings.Join(player.Queue(), " -> "))
	})

	commands.add(volume, func(e *gumble.TextMessageEvent) {
		
		if number, err := submatchExtract(reVolumeWithIn100, e.Message); player.HasStream() && err == nil {
			num, _ := strconv.ParseFloat(number, 32)
			player.SetVolume(float32(num/100))
		}

		sendMessage(e, fmt.Sprintf("Volume: %v%%\n", player.NormalizedVolume()))
	})

	commands.add(search, func(e *gumble.TextMessageEvent) {
		sendMessage(e, "Aqui saldran las busquedas de youtube")
	})
	
	commands.add(playSong, func(e *gumble.TextMessageEvent) {
		send := func(status string) {
			sendMessage(e, status)
		}
		if link, err := submatchExtract(rePlaySongHref, e.Message); !check(err) { 
			player.PlayOrQueue(link, send)
		}
		if search, err := submatchExtract(rePlaySongSearch, e.Message); !check(err) {
			player.PlayOrQueue(fmt.Sprintf("ytsearch1:%s", search), send)
		}
	})

	commands.add(playPause, func(e *gumble.TextMessageEvent) {
		player.PlayPause(func(status string) {
			sendMessage(e, status)
		})
	})

	commands.add(stop, func(e *gumble.TextMessageEvent) {
		player.Stop(func(status string) {
			sendMessage(e, status)
		})
	})

	commands.add(skip, func(e *gumble.TextMessageEvent) {
		player.Skip()
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
			player.SetClient(e.Client)
			player.Handlers()
			log.Println("Connected to the server")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil { return }
			if command, err := submatchExtract(reCommand, e.Message); !check(err) {
				commands.execute(command, e)
			}
			
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