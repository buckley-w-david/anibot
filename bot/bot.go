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
	discord.AddHandler(reactionAdd)

	// Open the websocket and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	defer discord.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("TDT Anibot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	m, _ := s.ChannelMessage(r.ChannelID, r.MessageID)
	field, err := tools.Reaction(*m, r.Emoji.Name)
	if err != nil {
		fmt.Println("Reaction now found: ", err)
		return
	}
	name := strings.TrimSuffix(field.Name, " "+r.Emoji.Name)

	switch name {
	case "Director":
		fmt.Println("Director: ", field.Value)
	case "Original Creator":
		fmt.Println("Creator: ", field.Value)
	case "Studio":
		fmt.Println("Studio: ", field.Value)
	default:
		fmt.Println("?: ", field.Value)
	}

	ctx := context.Background()
	media, err := anilist.MediaFromID(ctx, 11061)
	embed, _ := tools.Embed(media)
	s.ChannelMessageSendEmbed(m.ChannelID, &embed)
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
		fmt.Println(match)
		var requestType anilist.MediaType
		switch match[0] {
		case '{':
			requestType = anilist.ANIME
		case '<':
			requestType = anilist.MANGA
		}
		match = strings.TrimLeft(match, requestType.Left())
		match = strings.TrimRight(match, requestType.Right())
		media, err := requestType.MediaFromTitle(ctx, match, 1)
		if err != nil {
			fmt.Println("Error getting Media", err)
			return
		}
		tools.Send(s, m.ChannelID, media[0])
	}

}

func botCommand(s *discordgo.Session, channel string, request string) error {
	fmt.Println(request)

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
			media, err = anilist.MediaFromID(ctx, i)
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
		}
	case "director":
		fmt.Println("director")
	case "studio":
		fmt.Println("studio")
	default:
		fmt.Println("default")
	}
	return tools.Send(s, channel, media)
}
