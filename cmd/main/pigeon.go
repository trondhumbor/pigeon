package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/trondhumbor/pigeon/internal/command/serveralive"
	"github.com/trondhumbor/pigeon/internal/command/serverlist"
	"github.com/trondhumbor/pigeon/internal/server"
)

// CommandCreators is the list of handlers of the commands that are active
var CommandCreators = []server.CreateCommand{
	serveralive.CreateCommand,
	serverlist.CreateCommand,
}

func main() {

	configpath := flag.String(
		"config",
		"",
		"path to the config file")

	deletecommands := flag.Bool(
		"deletecommands",
		false,
		"if true, the program will delete all guild commands on startup, and then exit")

	flag.Parse()

	log.Println("starting pigeon")
	srv, err := server.New(*configpath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("creating discord session")
	sess := session.New("Bot " + srv.Token)

	log.Println("adding handlers and intents")
	sess.AddHandler(srv.MessageCreateHandler)
	sess.AddHandler(srv.HandleInteraction)
	// sess.AddHandler(srv.HandleComponentInteraction)

	sess.AddIntents(gateway.IntentGuilds)
	sess.AddIntents(gateway.IntentGuildMessages)
	sess.AddIntents(gateway.IntentGuildMessageReactions)
	sess.AddIntents(gateway.IntentDirectMessages)
	sess.AddIntents(gateway.IntentGuildMessageReactions)

	log.Println("opening discord session")
	err = sess.Open(context.Background())
	if err != nil {
		log.Fatalf("error opening connection: %v", err)
		return
	}

	defer sess.Close()

	log.Println("initializing server")
	err = srv.Initialize(sess, CommandCreators)
	if err != nil {
		log.Fatalf("error initializing server: %v", err)
		return
	}

	if deletecommands != nil && *deletecommands {
		log.Println("deletecommands flag detected, deleting guild commands...")
		err = srv.DeleteGuildCommands()
		if err != nil {
			log.Fatalf("error deleting commands: %v", err)
		}

		log.Println("deleted commands, exiting...")
		os.Exit(0)
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sigRec := <-sc

	log.Printf("signal %v received, exiting...", sigRec)
}
