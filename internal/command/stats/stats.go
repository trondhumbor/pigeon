package stats

import (
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

type statsHandler struct {
	session   *session.Session
	server    *server.Server
	formatter stringformat.Formatter
}

// CreateCommand creates a SlashCommand which handles /stats
func CreateCommand(srv *server.Server) (cmd command.SlashCommand, err error) {
	sh := statsHandler{session: srv.Session, server: srv, formatter: stringformat.New(srv.Mapnames, srv.Gametypes)}

	choices := []discord.StringChoice{}
	for _, master := range srv.MasterServers {
		gameId := master.GameId
		choices = append(choices, discord.StringChoice{Name: gameId, Value: gameId})
	}

	cmd = command.SlashCommand{
		HandleInteraction: sh.handleInteraction,
		CommandData: api.CreateCommandData{
			Name:        "stats",
			Description: "lists stats for the given game",
			Options: []discord.CommandOption{
				&discord.StringOption{
					OptionName:  "game",
					Description: "which game to list stats for",
					Required:    true,
					Choices:     choices,
				},
			},
		},
	}

	return
}

func (sh *statsHandler) handleInteraction(
	event *gateway.InteractionCreateEvent, options map[string]discord.CommandInteractionOption,
) (
	response *api.InteractionResponseData, err error,
) {
	var r string
	if servers, present := sh.server.GameServers[options["game"].String()]; present {
		var totalservers, totalplayers, totalbots int
		for _, s := range servers {
			c, cerr := strconv.Atoi(s["clients"])
			b, berr := strconv.Atoi(s["bots"])
			if cerr != nil || berr != nil {
				continue
			}

			// basic sanity checks
			if b > c || (c > 18 || c < 0) || (b > 18 || b < 0) {
				continue
			}

			totalplayers += c - b
			totalbots += b
			totalservers += 1
		}
		r = sh.formatter.Stats(totalservers, totalplayers, totalbots)
	} else {
		r = "couldn't find specified game in cache"
	}

	response = &api.InteractionResponseData{
		Content: option.NewNullableString(r),
	}
	return
}
