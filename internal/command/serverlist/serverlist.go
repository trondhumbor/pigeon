package serverlist

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/trondhumbor/pigeon/internal/command"
	"github.com/trondhumbor/pigeon/internal/server"
	"github.com/trondhumbor/pigeon/internal/stringformat"
)

type serverlistHandler struct {
	session *session.Session
	server  *server.Server
}

// CreateCommand creates a SlashCommand which handles /serverlist
func CreateCommand(srv *server.Server) (cmd command.SlashCommand, err error) {
	sh := serverlistHandler{session: srv.Session, server: srv}

	cmd = command.SlashCommand{
		HandleInteraction: sh.handleInteraction,
		CommandData: api.CreateCommandData{
			Name:        "serverlist",
			Description: "lists the servers for the given game",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "game",
					Description: "which game to show servers for",
					Required:    true,
					Choices: []discord.StringChoice{
						// StringChoice value must match MasterServer.gameId
						{Name: "h1", Value: "H1"},
						{Name: "iw6x", Value: "IW6"},
						{Name: "s1x", Value: "S1"},
					},
				},
				&discord.BooleanOption{
					OptionName:  "mobile",
					Description: "format serverlist for mobile devices",
					Required:    false,
				},
			},
		},
	}

	return
}

func (sh *serverlistHandler) handleInteraction(
	event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption,
) (
	response *api.InteractionResponseData, err error,
) {
	if servers, present := sh.server.GameServers[options["game"].String()]; present {
		desc := stringformat.DesktopList(servers)
		if val, present := options["mobile"]; present {
			mobile, err := val.BoolValue()
			if err != nil {
				mobile = false
			}
			if mobile {
				desc = stringformat.MobileList(servers)
			}
		}
		response = &api.InteractionResponseData{
			Content: option.NewNullableString(desc),
		}
		return
	} else {
		response = &api.InteractionResponseData{
			Content: option.NewNullableString("couldn't find specified game in cache"),
		}
		return
	}
}
