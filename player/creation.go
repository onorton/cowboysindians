package player

import "github.com/onorton/cowboysindians/message"

func CreatePlayer() *Player {
	name := message.RequestInput("Who are you?")
	return newPlayer(name)
}
