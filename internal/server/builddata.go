package server

import (
	"strings"
	"time"

	"github.com/trondhumbor/pigeon/internal/query"
)

func (srv *Server) PopulateGameServers() {

	querySingleServer := func(gameServer string, gameId string) {
		info, err := query.GetSingleServerResponse(gameServer)
		if err != nil {
			return
		}

		if val, present := info["gamename"]; present {
			if !strings.EqualFold(val, gameId) { // if server is not actually of the game we want
				return
			}
		} else {
			return // if gamename key isn't present
		}

		srv.gameServersWriteMutex.Lock()
		srv.GameServers[gameId] = append(srv.GameServers[gameId], info)
		srv.gameServersWriteMutex.Unlock()
	}

	query := func(master MasterServer) {
		servers := query.GetMasterServerResponse(master.Endpoint, master.GameId, master.GameVersion)

		for _, server := range servers {
			go querySingleServer(server, master.GameId)
		}
	}

	populate := func() {
		// create the initial arrays
		for _, m := range srv.MasterServers {
			srv.gameServersWriteMutex.Lock()
			srv.GameServers[m.GameId] = make([]GameServer, 0)
			srv.gameServersWriteMutex.Unlock()
		}

		for _, m := range srv.MasterServers {
			go query(m)
		}
	}

	// fill the cache initially
	populate()

	// refresh it every 3 minutes
	tickRate := 3 * time.Minute
	ticker := time.NewTicker(tickRate)
	go func() {
		for {
			select {
			case <-ticker.C:
				populate()
			}
		}
	}()
}
