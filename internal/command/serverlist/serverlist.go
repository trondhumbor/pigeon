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
	session   *session.Session
	server    *server.Server
	formatter stringformat.Formatter
}

// CreateCommand creates a SlashCommand which handles /serverlist
func CreateCommand(srv *server.Server) (cmd command.SlashCommand, err error) {
	sh := serverlistHandler{session: srv.Session, server: srv, formatter: stringformat.New(srv.Mapnames, srv.Gametypes)}

	choices := []discord.StringChoice{}
	for _, master := range srv.MasterServers {
		gameId := master.GameId
		choices = append(choices, discord.StringChoice{Name: gameId, Value: gameId})
	}

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
					Choices:     choices,
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

func (sh *serverlistHandler) sendMessage(event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption) {
	if servers, present := sh.server.GameServers[options["game"].String()]; present {
		if len(servers) == 0 {
			sh.session.SendMessage(event.ChannelID, "no servers found for the specified game.")
			return
		}

		desc := sh.formatter.DesktopList(servers)
		if val, present := options["mobile"]; present {
			mobile, err := val.BoolValue()
			if err != nil {
				mobile = false
			}
			if mobile {
				desc = sh.formatter.MobileList(servers)
			}
		}
		for _, m := range desc {
			sh.session.SendMessage(event.ChannelID, m)
		}
	} else {
		sh.session.SendMessage(event.ChannelID, "couldn't find specified game in cache")
	}
}

func (sh *serverlistHandler) handleInteraction(
	event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption,
) (
	response *api.InteractionResponseData, err error,
) {
	go sh.sendMessage(event, options)

	response = &api.InteractionResponseData{
		Content: option.NewNullableString("sending server list, please wait..."),
	}
	return
}
