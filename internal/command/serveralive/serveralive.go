package serveralive

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/trondhumbor/pigeon/internal/command"
	"github.com/trondhumbor/pigeon/internal/server"
	"github.com/trondhumbor/pigeon/internal/stringformat"
)

type serveraliveHandler struct {
	session   *session.Session
	server    *server.Server
	formatter stringformat.Formatter
}

// CreateCommand creates a SlashCommand which handles /serveralive
func CreateCommand(srv *server.Server) (cmd command.SlashCommand, err error) {
	sh := serveraliveHandler{session: srv.Session, server: srv, formatter: stringformat.New(srv.Mapnames, srv.Gametypes)}

	cmd = command.SlashCommand{
		HandleInteraction: sh.handleInteraction,
		CommandData: api.CreateCommandData{
			Name:        "serveralive",
			Description: "lists the servers for the given game",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "filter",
					Description: "filter for servers with hostname",
					Required:    true,
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

func (sh *serveraliveHandler) sendMessage(event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption) {
	var servers []server.GameServer
	for _, v := range sh.server.GameServers {
		servers = append(servers, v...)
	}

	if val, present := options["filter"]; present {
		servers = filter(servers, val.String())
	}

	if len(servers) == 0 {
		sh.session.SendMessage(event.ChannelID, "no servers found for the specified filter.")
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
}

func filter(list []server.GameServer, filterString string) []server.GameServer {
	var ret []server.GameServer
	for _, s := range list {
		if val, present := s["hostname"]; present {
			if strings.Contains(strings.ToLower(val), strings.ToLower(filterString)) {
				ret = append(ret, s)
			}
		}
	}
	return ret
}

func (sh *serveraliveHandler) handleInteraction(
	event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption,
) (
	response *api.InteractionResponseData, err error,
) {
	go sh.sendMessage(event, options)

	response = &api.InteractionResponseData{
		Content: option.NewNullableString("sending filtered server list, please wait..."),
	}
	return
}
