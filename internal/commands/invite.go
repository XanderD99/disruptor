package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/XanderD99/disruptor/internal/disruptor"
)

type invite struct{}

func Invite() disruptor.Command {
	return invite{}
}

// Load implements disruptor.Command.
func (i invite) Load(r handler.Router) {
	r.SlashCommand("/invite", i.handle)
}

// Options implements disruptor.Command.
func (i invite) Options() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "invite",
		Description: "Get an invite link to add Disruptor to your server",
	}
}

func (i invite) handle(_ discord.SlashCommandInteractionData, event *handler.CommandEvent) error {
	inviteLink := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&permissions=39584569298176&scope=bot%%20applications.commands",
		event.Client().ID(),
	)

	content := fmt.Sprintf("Invite Disruptor to your server: [click me](%s)", inviteLink)

	msg := discord.NewMessageUpdateBuilder().SetContent(content).Build()

	_, err := event.UpdateInteractionResponse(msg)
	return err
}

var _ disruptor.Command = (*invite)(nil)
