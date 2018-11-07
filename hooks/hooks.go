package main

import (
	"context"
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
	hookURL string
)

func main() {
	hookURL, _ = shared.Hook.OrEnv()
	botToken, err := shared.Token.OrEnv()

	if err != nil {
		fmt.Println(shared.MissingToken)
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

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":1236", nil))
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
}

func debug(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println(m.ChannelID)
}
