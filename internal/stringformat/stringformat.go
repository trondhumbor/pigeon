package stringformat

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/trondhumbor/pigeon/internal/server"
)

func leftjust(s string, n int) string {
	i := 0
	if len(s) < n {
		i = n - len(s)
	} else {
		return s[:n]
	}

	return s + strings.Repeat(" ", i)
}

func sanitizeFields(inServer server.GameServer) server.GameServer {
	sanitized := make(server.GameServer)
	re := regexp.MustCompile(`(?i)^https?:\/\/`) // remove links that discord tries to parse
	for k, v := range inServer {
		sanitized[k] = strings.ReplaceAll(v, "`", "")
		sanitized[k] = re.ReplaceAllString(v, "hxxp://")
	}
	return sanitized
}

type Formatter struct {
	mapnames  map[string]string
	gametypes map[string]string
}

func New(mapnames, gametypes map[string]string) (f Formatter) {
	return Formatter{mapnames: mapnames, gametypes: gametypes}
}

func (f *Formatter) MapnameLookup(key string) string {
	if val, present := f.mapnames[key]; present {
		return val
	}
	return "Unknown" // could potentially return the key here instead
}

func (f *Formatter) GametypeLookup(key string) string {
	if val, present := f.gametypes[key]; present {
		return val
	}
	return "Unknown" // could potentially return the key here instead
}

func (f *Formatter) DesktopList(servers []server.GameServer) []string {
	var messages []string
	desc := "```\n"
	for _, s := range servers {
		// if the next server will exceed the discord char limit, cut it off and start on a new message
		if len(desc)+100 > 2000 {
			desc += "```"
			messages = append(messages, desc)
			desc = "```\n"
		}

		s = sanitizeFields(s)
		hostname := leftjust(s["hostname"], 37)
		mapname := leftjust(f.MapnameLookup(s["mapname"]), 17)
		gametype := leftjust(f.GametypeLookup(s["gametype"]), 10)
		clients := leftjust(fmt.Sprintf("%s / %s", s["clients"], s["sv_maxclients"]), 7)
		desc += fmt.Sprintf("| %s | %s | %s | %s |\n", hostname, mapname, gametype, clients)
	}
	desc += "```"
	messages = append(messages, desc)
	return messages
}

func (f *Formatter) MobileList(servers []server.GameServer) []string {
	var messages []string
	desc := "```\n---------------------------------\n"
	for _, s := range servers {
		// if the next server will exceed the discord char limit, cut it off and start on a new message
		if len(desc)+200 > 2000 {
			desc += "```"
			messages = append(messages, desc)
			desc = "```\n---------------------------------\n"
		}

		s = sanitizeFields(s)
		hostname := fmt.Sprintf("|%s|%s|", leftjust("Hostname", 8), leftjust(s["hostname"], 22))
		mapname := fmt.Sprintf("|%s|%s|", leftjust("Map", 8), leftjust(f.MapnameLookup(s["mapname"]), 22))
		gametype := fmt.Sprintf("|%s|%s|", leftjust("Gametype", 8), leftjust(f.GametypeLookup(s["gametype"]), 22))
		clients := fmt.Sprintf("|%s|%s|", leftjust("Clients", 8), leftjust(fmt.Sprintf("%s / %s", s["clients"], s["sv_maxclients"]), 22))
		desc += fmt.Sprintf("%s\n%s\n%s\n%s\n", hostname, mapname, gametype, clients)
		desc += "---------------------------------\n"
	}
	desc += "```"
	messages = append(messages, desc)
	return messages
}
