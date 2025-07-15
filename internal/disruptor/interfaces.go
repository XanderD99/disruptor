package disruptor

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

type Command interface {
	Load(router handler.Router)
	Options() discord.SlashCommandCreate
}
