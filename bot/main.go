package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/bwmarrin/discordgo"
	//"github.com/davecgh/go-spew/spew"
)

var (
	token string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func embed(media anilist.MediaResponse) discordgo.MessageEmbed {
	coverImage := discordgo.MessageEmbedThumbnail{
		URL: media.Media.CoverImage.Medium,
	}

	mediaType := discordgo.MessageEmbedField{
		Name:   "Media Type",
		Value:  fmt.Sprintf("%s %s %s", media.Media.MediaType, media.Media.Format, media.Media.Source),
		Inline: false,
	}
	fields := []*discordgo.MessageEmbedField{&mediaType}

	studios := make([]*discordgo.MessageEmbedField, len(media.Media.Studios.Edges))
	for i, studio := range media.Media.Studios.Edges {
		studios[i] = &discordgo.MessageEmbedField{
			Name:   "Studio",
			Value:  fmt.Sprintf("[%s](%s)", studio.Node.Name, studio.Node.SiteURL),
			Inline: false,
		}
	}
	fields = append(fields, studios...)

	directorName, directorURL, directorErr := media.Director()

	var director discordgo.MessageEmbedField
	if directorErr == nil {
		director = discordgo.MessageEmbedField{
			Name:   "Director",
			Value:  fmt.Sprintf("[%s](%s)", directorName, directorURL),
			Inline: true,
		}
		fields = append(fields, &director)
	}

	creatorName, creatorURL, creatorErr := media.Creator()

	var creator discordgo.MessageEmbedField
	if creatorErr == nil {
		creator = discordgo.MessageEmbedField{
			Name:   "Original Creator",
			Value:  fmt.Sprintf("[%s](%s)", creatorName, creatorURL),
			Inline: true,
		}
		fields = append(fields, &creator)
	}

	return discordgo.MessageEmbed{
		URL:         media.Media.SiteURL,
		Title:       media.Media.Title.Romaji,
		Description: media.Media.Description,
		Color:       0x00ff00,
		Thumbnail:   &coverImage,
		Fields:      fields,
	}
}

func main() {
	if token == "" {
		token = os.Getenv("TOKEN")
		if token == "" {
			fmt.Println("No token provided. Please run: tdt-anibot -t <bot token>")
			return
		}
	}

	// Create a new Discord session using the provided bot token.
	discord, err := discordgo.New("Bot " + token)
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
		err = discord.UpdateStatusComplex(status)

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
	}
	defer discord.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("TDT Anibot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
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
				res, err := anilist.MediaFromID(ctx, anilist.IDQuery{ID: i, Type: queryType})
				if err != nil {
					fmt.Println("Error getting Media", err)
					return
				}
				embed := embed(res)
				s.ChannelMessageSendEmbed(m.ChannelID, &embed)
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
				embed := embed(res)
				s.ChannelMessageSendEmbed(m.ChannelID, &embed)
			}
		case "director":
			fmt.Println("three")
		case "studio":
			fmt.Println("three")
		default:
			fmt.Println("three")
		}
	}
}
