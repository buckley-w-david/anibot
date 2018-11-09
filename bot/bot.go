package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/anibot/tools"
	"github.com/bwmarrin/discordgo"
	//"github.com/davecgh/go-spew/spew"
)

var (
	discord *discordgo.Session

	re *regexp.Regexp
)

func init() {
	tools.SetupSharedOptions()
	flag.Parse()

	re = regexp.MustCompile("({.*?})|(<.*?>)")
}

func main() {
	botToken, err := tools.Token.OrEnv()
	if err != nil {
		fmt.Println(tools.MissingToken)
		return
	}

	// Create a new Discord session using the provided bot token.
	discord, err = discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	// Register ready as a callback for the ready events.
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		game := discordgo.Game{
			Name: "anilist",
			Type: discordgo.GameTypeWatching,
		}

		status := discordgo.UpdateStatusData{
			Game: &game,
			AFK:  false,
		}
		err := discord.UpdateStatusComplex(status)

		if err != nil {
			fmt.Println("Error attempting to set my status")
		}
	})

	// Register messageCreate as a callback for the messageCreate events.
	discord.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	defer discord.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Anibot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Service direct bot commands
	if strings.HasPrefix(m.Content, "!anibot ") {
		request := strings.TrimPrefix(m.Content, "!anibot ")
		err := botCommand(s, m.ChannelID, request)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	// Parse message for any implicit anime/manga requests
	ctx := context.Background()
	for _, match := range re.FindAllString(m.Content, -1) {
		var requestType anilist.MediaType
		switch match[0] {
		case '{':
			requestType = anilist.ANIME
			match = strings.TrimLeft(match, "}")
			match = strings.TrimRight(match, "{")
		case '<':
			requestType = anilist.MANGA
			match = strings.TrimLeft(match, ">")
			match = strings.TrimRight(match, "<")
		}

		query := anilist.MediaQuery{
			Title:      match,
			Type:       requestType.String(),
			MaxResults: 1,
		}
		media, err := anilist.MediaFromMediaQuery(ctx, query)
		if err != nil {
			fmt.Println("Error getting Media", err)
			return
		}
		err = tools.Send(s, m.ChannelID, media[0])
		if err != nil {
			fmt.Println("Error sending message", err)
		}
	}
}

func botCommand(s *discordgo.Session, channel string, request string) error {
	r := csv.NewReader(strings.NewReader(request))
	r.Comma = ' ' // space
	fields, err := r.Read()
	if err != nil {
		return err
	}
	command := fields[0]
	args := fields[1:len(fields)]

	var queryType string
	if strings.ToLower(args[0]) == "manga" {
		queryType = "MANGA"
		args = args[1:len(args)]
	} else if strings.ToLower(args[0]) == "anime" {
		queryType = "ANIME"
		args = args[1:len(args)]
	} else {
		queryType = ""
	}

	var media anilist.Media
	switch command {
	case "id":
		for _, id := range args {
			i, err := strconv.Atoi(id)
			if err != nil {
				return err
			}

			ctx := context.Background()
			media, err = anilist.MediaFromMediaID(ctx, i)
			if err != nil {
				return err
			}
		}
	case "title":
		for _, title := range args {
			ctx := context.Background()
			medias, err := anilist.MediaFromMediaQuery(ctx, anilist.MediaQuery{Title: title, Type: queryType, MaxResults: 1})
			if err != nil {
				return err
			}
			media = medias[0]
			tools.Send(s, channel, media)
		}
	case "person":
		for _, name := range args {
			ctx := context.Background()
			medias, err := anilist.MediaFromPersonQuery(ctx, anilist.PersonQuery{Name: name, Type: queryType, MaxResults: 1})
			if err != nil {
				return err
			}
			media = medias[0]
			tools.Send(s, channel, media)
		}
	case "studio":
		for _, name := range args {
			ctx := context.Background()
			medias, err := anilist.MediaFromStudioQuery(ctx, anilist.StudioQuery{Name: name, MaxResults: 1})
			if err != nil {
				return err
			}
			media = medias[0]
			tools.Send(s, channel, media)
		}
	default:
		fmt.Println("default")
	}
	return nil
}
