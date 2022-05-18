package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/trondhumbor/pigeon/internal/command"
)

// CreateCommand is a function that returns a list of SlashCommands
type CreateCommand func(*Server) (command.SlashCommand, error)
type GameServer = map[string]string

type MasterServer struct {
	GameId   string `json:"gameId"`
	Protocol int    `json:"protocol"`
	Endpoint string `json:"endpoint"`
}

// Server is the config and main server
type Server struct {
	Token         string          `json:"token"`
	AppID         discord.AppID   `json:"appID"`
	GuildID       discord.GuildID `json:"guildID"`
	MasterServers []MasterServer  `json:"masterServers"`

	commands map[string]command.SlashCommand

	Session               *session.Session
	LastMessages          map[discord.ChannelID]*gateway.MessageCreateEvent
	lastMessageWriteMutex sync.Mutex

	GameServers           map[string][]GameServer
	gameServersWriteMutex sync.Mutex
}

// New creates a new server instance with initialized variables
func New(configpath string) (srv Server, err error) {
	srv = Server{
		LastMessages: make(map[discord.ChannelID]*gateway.MessageCreateEvent),
		GameServers:  make(map[string][]GameServer),
	}

	log.Printf("reading config file from %q", configpath)
	f, err := ioutil.ReadFile(configpath)
	if err != nil {
		log.Printf("failed to read file: %v", err)
		return
	}

	err = json.Unmarshal(f, &srv)
	if err != nil {
		log.Printf("failed to unmarshall file: %v", err)
		return
	}

	srv.commands = map[string]command.SlashCommand{}

	return
}

// Initialize the server with the given session
func (srv *Server) Initialize(s *session.Session, commandCreators []CreateCommand) error {
	srv.Session = s

	log.Printf("creating/updating %d guild commands...", len(commandCreators))

	cmdMap := make(map[string]command.SlashCommand)
	cmdList := []api.CreateCommandData{}

	for _, createCMD := range commandCreators {
		cmd, err := createCMD(srv)
		if err != nil {
			return fmt.Errorf("error creating command: %v", err)
		}

		cmdMap[cmd.CommandData.Name] = cmd
		cmdList = append(cmdList, cmd.CommandData)
	}

	_, err := s.BulkOverwriteGuildCommands(srv.AppID, srv.GuildID, cmdList)
	if err != nil {
		return fmt.Errorf("bulk overwrite guild commands: %v", err)
	}

	go srv.PopulateGameServers()

	srv.commands = cmdMap
	return nil
}

// MessageCreateHandler handles every incoming normal message
func (srv *Server) MessageCreateHandler(c *gateway.MessageCreateEvent) {
	srv.lastMessageWriteMutex.Lock()
	srv.LastMessages[c.ChannelID] = c
	srv.lastMessageWriteMutex.Unlock()
}

// HandleInteraction is a handler-function handling interaction-events
func (srv *Server) HandleInteraction(ev *gateway.InteractionCreateEvent) {
	switch ev.Data.(type) {
	case *discord.CommandInteraction:
		data := ev.Data.(*discord.CommandInteraction)
		srv.handleCommandInteraction(ev, data)
		return
	}
}

//revive:disable-next-line:cyclomatic
// handleCommandInteraction is a handler-function handling interaction-events
func (srv *Server) handleCommandInteraction(
	event *gateway.InteractionCreateEvent,
	data *discord.CommandInteraction,
) {

	options, err := opsToMap(data.Options)
	if err != nil {
		log.Printf("error occurred converting ops to a map: %v", err)
		return
	}

	cmd, exists := srv.commands[data.Name]
	if !exists {
		log.Printf("command %s does not exist", data.Name)
		return
	}

	responseData, err := cmd.HandleInteraction(event, options)
	if err != nil {
		log.Printf("error occurred handling interaction: %v", err)
		dm, dmErr := srv.Session.CreatePrivateChannel(event.Member.User.ID)
		if dmErr != nil {
			log.Printf("error occurred creating private channel to report error: %v", dmErr)
			return
		}

		_, dmErr = srv.Session.SendMessage(dm.ID, err.Error())
		if dmErr != nil {
			log.Printf("error occurred sending DM to report error: %v", dmErr)
			return
		}

		return
	}

	interactionResp := api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: responseData,
	}
	if err := srv.Session.RespondInteraction(event.ID, event.Token, interactionResp); err != nil {
		log.Printf("failed to send interaction callback: %v", err)
		return
	}

	log.Printf("responded to interaction")
}

// DeleteGuildCommands deletes all guild commands for the configured guild and app ID
func (srv *Server) DeleteGuildCommands() error {
	cmds, err := srv.Session.GuildCommands(srv.AppID, srv.GuildID)
	if err != nil {
		return fmt.Errorf("fetching existing commands: %v", err)
	}

	for i, cmd := range cmds {
		err = srv.Session.DeleteGuildCommand(
			cmd.AppID,
			srv.GuildID,
			cmd.ID)
		if err != nil {
			return fmt.Errorf("deleting command %s: %v", cmd.Name, err)
		}
		log.Printf("deleted guild command (%d/%d) %q", i+1, len(cmds), cmd.Name)
	}

	return nil
}

func opsToMap(ops discord.CommandInteractionOptions) (
	opMap map[string]discord.CommandInteractionOption,
	err error,
) {
	opMap = make(map[string]discord.CommandInteractionOption)
	for _, op := range ops {
		opMap[op.Name] = op
	}

	return
}
