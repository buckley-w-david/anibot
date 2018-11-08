package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/anibot/shared"
	"github.com/bwmarrin/discordgo"
	//"github.com/davecgh/go-spew/spew"
)

var (
	discord *discordgo.Session
)

func init() {
	shared.SetupSharedOptions()
	flag.Parse()
}

func main() {
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
	field, err := shared.Reaction(*m, r.Emoji.Name)
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
	embed, _ := shared.Embed(media)
	s.ChannelMessageSendEmbed(m.ChannelID, &embed)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!anibot ") {
		request := strings.TrimPrefix(m.Content, "!anibot ")
		r := csv.NewReader(strings.NewReader(request))
		r.Comma = ' ' // space
		fields, err := r.Read()
		if err != nil {
			fmt.Println(err)
			return
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

		var embed discordgo.MessageEmbed
		switch command {
		case "id":
			for _, id := range args {
				i, err := strconv.Atoi(id)
				fmt.Println(i)
				if err != nil {
					fmt.Println("Error parsing request", err)
					return
				}
				ctx := context.Background()
				res, err := anilist.MediaFromID(ctx, i)
				if err != nil {
					fmt.Println("Error getting Media", err)
					return
				}

				embed, err = shared.Embed(res)
				if err != nil {
					log.Println(err)
					return
				}

			}
		case "title":
			for _, title := range args {
				fmt.Println(title)
				ctx := context.Background()
				res, err := anilist.MediaFromTitle(ctx, anilist.TitleQuery{Title: title, Type: queryType})
				if err != nil {
					fmt.Println("Error getting Media", err)
					return
				}

				embed, err = shared.Embed(res)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "director":
			fmt.Println("director")
		case "studio":
			fmt.Println("studio")
		default:
			fmt.Println("default")
		}
		sent, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
		for _, emoji := range shared.Emojis() {
			s.MessageReactionAdd(sent.ChannelID, sent.ID, emoji)
		}
	}
}
