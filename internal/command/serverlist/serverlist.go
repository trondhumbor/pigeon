package serverlist

import (
	"log"
	"strconv"

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
				&discord.BooleanOption{
					OptionName:  "full",
					Description: "show full servers",
					Required:    false,
				},
				&discord.BooleanOption{
					OptionName:  "empty",
					Description: "show empty servers",
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
			_, mErr := sh.session.SendMessage(event.ChannelID, "no servers found for the specified game.")
			if mErr != nil {
				log.Printf("error occured sending message")
			}
			return
		}

		servers := filter(servers, options)

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
			_, mErr := sh.session.SendMessage(event.ChannelID, m)
			if mErr != nil {
				log.Printf("error occured sending message")
			}
		}
	} else {
		_, mErr := sh.session.SendMessage(event.ChannelID, "couldn't find specified game in cache")
		if mErr != nil {
			log.Printf("error occured sending message")
		}
	}
}

func filter(list []server.GameServer, options map[string]discord.CommandInteractionOption) []server.GameServer {
	full := true
	empty := true
	var err error

	if val, present := options["full"]; present {
		full, err = val.BoolValue()
		if err != nil {
			full = true
		}
	}

	if val, present := options["empty"]; present {
		empty, err = val.BoolValue()
		if err != nil {
			empty = true
		}
	}

	var ret []server.GameServer

	for _, s := range list {
		c, cerr := strconv.Atoi(s["clients"])
		b, berr := strconv.Atoi(s["bots"])
		m, merr := strconv.Atoi(s["sv_maxclients"])
		if cerr != nil || berr != nil || merr != nil {
			continue
		}

		// basic sanity checks
		if b > c || (c > 18 || c < 0) || (b > 18 || b < 0) || (m > 18 || m < 0) {
			continue
		}

		if !full && c == m {
			continue
		}

		if !empty && c == 0 {
			continue
		}

		ret = append(ret, s)
	}

	return ret
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
