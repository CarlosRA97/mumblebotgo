package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"
	_ "layeh.com/gumble/opus"
)

const (
	PLAY = "play"
	PAUSE = "pause"
	STOP = "stop"
)

var (
	commandFunc = make(map[string]func(*gumble.TextMessageEvent))
	availableCommands = []string{PLAY, PAUSE, STOP}
)

func main() {
	var stream *gumbleffmpeg.Stream
	var offset time.Duration

	reCommand, err := regexp.Compile(fmt.Sprintf("(?:!)(%s)*", strings.Join(availableCommands, "|")))
	if check(err) { return }
	reHref, err := regexp.Compile("(?:\")(.*)(?:\")")
	if check(err) { return }

	submatchExtract := func (match [][]string) (string, error) {
		err := errors.New("No match")
		fmt.Println(match)
		if len(match) > 0 && len(match[0]) > 1 {
			return match[0][1], nil
		}
		return "", err
	}

	commandFunc[PLAY] = func(e *gumble.TextMessageEvent) {
		streamStateInitial := stream != nil && (
			stream.State() == 0	||
			stream.State() == gumbleffmpeg.StateInitial )
		streamStatePaused := stream != nil && 
			stream.State() == gumbleffmpeg.StatePaused

		if stream == nil {
			matchesLink := reHref.FindAllStringSubmatch(e.Message, -1)
			link, err := submatchExtract(matchesLink)
			fmt.Println(link)
			downloaded := "downloaded.ogg"
			if check(err) { return }

			cmd := exec.Command("youtube-dl","--extract-audio","--audio-format","vorbis", link, "--output", fmt.Sprintf("./%s", downloaded))
			// "ytsearch1: " + e.Message[1:] + ""
			if err := cmd.Run(); err == nil {
				fmt.Printf("Descargado %s\n", link)
			} else {
				fmt.Printf("%s\n", err)
			}

			stream = gumbleffmpeg.New(e.Client, gumbleffmpeg.SourceFile(downloaded))
			// stream.Volume = 0.01
			if err := stream.Play(); err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Playing %s\n", link)
			}
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

	commandFunc[PAUSE] = func(e *gumble.TextMessageEvent) {
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

	commandFunc[STOP] = func(e *gumble.TextMessageEvent) {
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
				if err := os.Remove("downloaded.ogg"); err != nil {
					fmt.Println("No se pudo borrar el archivo downloaded.ogg")
				}
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
			fmt.Println("Connected to the server")
			os.Remove("downloaded.ogg")
		},

		TextMessage: func(e *gumble.TextMessageEvent) {
			if e.Sender == nil {
				return
			}

			// hasStream := stream != nil
			// streamStatePaused := hasStream && 
			// 		stream.State() == gumbleffmpeg.StatePaused
			// streamStateInitial := hasStream && (
			// 		stream.State() == 0	||
			// 		stream.State() == gumbleffmpeg.StateInitial )
			// streamStatePlaying := hasStream &&
			// 		stream.State() == gumbleffmpeg.StatePlaying

			// if hasStream {
			// 	fmt.Println(stream.State())
			// }

			matchesCommand := reCommand.FindAllStringSubmatch(e.Message, -1)
			fmt.Println(matchesCommand)		
		
			command, err := submatchExtract(matchesCommand)
			if check(err) { return }
			commandFunc[command](e)
			// isPlayCommand := err == nil && strings.Contains(command, PLAY)
			// isPauseCommand := err == nil && strings.Contains(command, PAUSE)
			// isStopCommand := err == nil && strings.Contains(command, STOP)

			// if !hasStream && isPlayCommand {
			// 	matchesLink := reHref.FindAllStringSubmatch(e.Message, -1)
			// 	link, err := submatchExtract(matchesLink)
			// 	fmt.Println(link)
			// 	downloaded := "downloaded.ogg"
			// 	if check(err) { return }

			// 	cmd := exec.Command("youtube-dl","--extract-audio","--audio-format","vorbis", link, "--output", fmt.Sprintf("./%s", downloaded))
			// 	// "ytsearch1: " + e.Message[1:] + ""
			// 	if err := cmd.Run(); err == nil {
			// 		fmt.Printf("Descargado %s\n", link)
			// 	} else {
			// 		fmt.Printf("%s\n", err)
			// 	}

			// 	stream = gumbleffmpeg.New(e.Client, gumbleffmpeg.SourceFile(downloaded))
			// 	// stream.Volume = 0.01
			// 	if err := stream.Play(); err != nil {
			// 		fmt.Printf("%s\n", err)
			// 	} else {
			// 		fmt.Printf("Playing %s\n", link)
			// 	}
			// 	return 
				
			// }

			// if (streamStateInitial || streamStatePaused) && isPlayCommand {
			// 	stream.Offset = offset
			// 	if err := stream.Play(); err != nil {
			// 		fmt.Printf("%s\n", err)
			// 	} else {
			// 		fmt.Printf("Playing\n")
			// 	}
			// 	return
			// }

			// fmt.Printf("Pause requirements: %v, %v\n", streamStatePlaying, isPauseCommand)
			// if hasStream && streamStatePlaying && isPauseCommand {
			// 	fmt.Println(e.Message)
			// 	if err := stream.Pause(); err != nil {
			// 		fmt.Printf("%s\n", err)
			// 	} else {
			// 		offset = stream.Offset
			// 		fmt.Printf("Pausing\n")
			// 	}
			// 	return
			// }


			// fmt.Printf("Stop requirements: %v, %v\n", streamStatePlaying, isStopCommand)
			// if (streamStatePlaying || streamStatePaused) && isStopCommand {
			// 	if err := stream.Stop(); err != nil {
			// 		fmt.Printf("%s\n", err)
			// 	} else {
			// 		fmt.Printf("Stopped\n")
			// 		if err := os.Remove("downloaded.ogg"); err != nil {
			// 			fmt.Println("No se pudo borrar el archivo downloaded.ogg")
			// 		}
			// 	}
			// 	return
			// }

			// switch matches[0][1:] {
			// case "play":
			// 	if play(stream, e.Client, e.Message, &offset) { return }
			// 	break
			// case "pause":
			// 	if pause(stream, e.Message, &offset) { return }
			// 	break
			// case "stop":
			// 	if stop(stream, e.Message) { return }
			// 	fmt.Println("Stopping")
			// 	break
			// }

			
			
			
			if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
				return
			}

			
			// downloaded := "downloaded.ogg"
			// cmd := exec.Command("youtube-dl","--extract-audio","--audio-format","vorbis", resource, "--output", fmt.Sprintf("./%s", downloaded))
			// if err := cmd.Run(); err == nil {
			// 	fmt.Printf("Descargado %s\n", resource)
			// } else {
			// 	fmt.Printf("%s\n", err)
			// }

			// stream := gumbleffmpeg.New(e.Client, gumbleffmpeg.SourceFile(downloaded))
			// if err := stream.Play(); err != nil {
			// 	fmt.Printf("%s\n", err)
			// } else {
			// 	fmt.Printf("Playing %s\n", resource)
			// }
		},
	})
}

func try(msg string, err error) string {
	if check(err) { return "" }
	return msg
}

func linkStripper(anchorTag string) string { return strings.Split(anchorTag, "\"")[1] }

// func play(stream *gumbleffmpeg.Stream, client *gumble.Client, message string, offset *time.Duration) bool {
// 	fmt.Println(strings.Contains(message, "https"))
// 	if stream != nil && (stream.State() == gumbleffmpeg.StatePaused || stream.State() == gumbleffmpeg.StateInitial || stream.State() == 0) && strings.Contains(matches[0][1:], "play") {
// 		if strings.Contains(message, "https") {
// 			stream = playSource(client, linkStripper(message))
// 		}

// 		stream.Offset = *offset
// 		if err := stream.Play(); err != nil {
// 			fmt.Printf("%s\n", err)
// 		} else {
// 			fmt.Printf("Playing\n")
// 		}
// 		return true
// 	}
// 	return false
// }

// func pause(stream *gumbleffmpeg.Stream, message string, offset *time.Duration) bool {
// 	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying && strings.Contains(message, "pause") {
// 		if err := stream.Pause(); err != nil {
// 			fmt.Printf("%s\n", err)
// 		} else {
// 			*offset = stream.Offset
// 			fmt.Printf("Pausing\n")
// 		}
// 		return true
// 	}
// 	return false
// }

// func stop(stream *gumbleffmpeg.Stream, message string) bool {
// 	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying && strings.Contains(message, "stop") {
// 		if err := stream.Stop(); err != nil {
// 			fmt.Printf("%s\n", err)
// 		} else {
// 			fmt.Printf("Stopped\n")
// 			if err := os.Remove("downloaded.ogg"); err != nil {
// 				fmt.Println("No se pudo borrar el archivo downloaded.ogg")
// 			}
// 		}
// 		return true
// 	}
// 	return false
// }

// func playSource(client *gumble.Client, resource string) *gumbleffmpeg.Stream {
// 	downloaded := "downloaded.ogg"
// 	cmd := exec.Command("youtube-dl","--extract-audio","--audio-format","vorbis", resource, "--output", fmt.Sprintf("./%s", downloaded))
// 	if err := cmd.Run(); err == nil {
// 		fmt.Printf("Descargado %s\n", resource)
// 	} else {
// 		fmt.Printf("%s\n", err)
// 	}

// 	stream := gumbleffmpeg.New(client, gumbleffmpeg.SourceFile(downloaded))
// 	if err := stream.Play(); err != nil {
// 		fmt.Printf("%s\n", err)
// 	} else {
// 		fmt.Printf("Playing %s\n", resource)
// 	}
// 	return stream
// }

func check(err error) bool {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return true
	}
	return false
}