package player

import (
	"log"
	"time"

	"layeh.com/gumble/gumbleffmpeg"
)

func (p *Player) Handlers() {
	go p.queueHandler()
	go p.usersPresenseHandler()
} 

func (p *Player) queueHandler() {
	for {
		if p.HasStream() {
			switch p.stream.State() {
				case gumbleffmpeg.StatePlaying: {
					p.stream.Wait()
					log.Println("He terminado la cancion")
					p.Skip()
					if source, err := p.Dequeue(); err == nil {
						log.Printf("Siguente cancion %s\n", source)
						p.Play(source)
					} else {
						p.Stop(nil)
						log.Println("No hay mas canciones en la cola")
					}
				}; break
			}
		}
		time.Sleep(time.Second * 1)
	}	
}

func (p *Player) usersPresenseHandler() {
	for {
		if p.HasStream() && len(p.client.Users) <= 1 {
			log.Println("Stopping: There is no one listening!")
			p.Stop(nil)
		}
		time.Sleep(time.Second * 30)
	}
}