package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/anibot/shared"
	"github.com/bwmarrin/discordgo"
	//"github.com/davecgh/go-spew/spew"
)

var (
	discord *discordgo.Session
	port    shared.CliOption
	hookURL string
)

func init() {
	port = shared.CliOption{Name: "port", Short: "p", Description: "Port to listen on", DefaultValue: "1236"}
	port.StringVar()
	shared.SetupSharedOptions()

	flag.Parse()
}

func main() {
	hookURL, _ = shared.Hook.OrEnv()
	botToken, err := shared.Token.OrEnv()
	if err != nil {
		fmt.Println(shared.MissingToken)
		return
	}

	listenPort, err := port.OrEnv()
	if err != nil {
		fmt.Println("No port specified")
		return
	}

	// Create a new Discord session using the provided bot token.
	discord, err = discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	// Open the websocket and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	defer discord.Close()
	discord.AddHandler(debug)

	http.HandleFunc("/hooks", handler)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	media, err := anilist.MediaFromID(ctx, 11061)
	if err != nil {
		fmt.Println("Error getting media", err)
		return
	}

	channels, ok := r.URL.Query()["channel"]
	if !ok || len(channels[0]) < 1 {
		log.Println("Url Param 'channel' is missing")
		return
	}
	channel := channels[0]

	var embed discordgo.MessageEmbed
	if hookURL != "" {
		embed, _ = shared.EmbedHook(media, channel, hookURL)
	} else {
		embed, _ = shared.Embed(media)
	}
	discord.ChannelMessageSendEmbed(channel, &embed)
	fmt.Fprint(w, "OK")
}

func debug(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.ChannelID)
}
