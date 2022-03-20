package command

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// SlashCommand is data about a command and the function to handle interactions in the way described
// by the data
type SlashCommand struct {
	CommandData       api.CreateCommandData
	HandleInteraction func(
		event *gateway.InteractionCreateEvent,
		options map[string]discord.CommandInteractionOption,
	) (*api.InteractionResponseData, error)
}
