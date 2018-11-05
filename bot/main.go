package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/buckley-w-david/tdt-anibot/anilist"
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
			Value:  fmt.Sprintf("[%s](%s)", studio.Node.Name, studio.Node.SiteUrl),
			Inline: false,
		}
	}
	fields = append(fields, studios...)

	directorErr, directorName, directorUrl := media.Director()

	var director discordgo.MessageEmbedField
	if directorErr == nil {
		director = discordgo.MessageEmbedField{
			Name:   "Director",
			Value:  fmt.Sprintf("[%s](%s)", directorName, directorUrl),
			Inline: true,
		}
		fields = append(fields, &director)
	}

	creatorErr, creatorName, creatorUrl := media.Creator()

	var creator discordgo.MessageEmbedField
	if creatorErr == nil {
		creator = discordgo.MessageEmbedField{
			Name:   "Original Creator",
			Value:  fmt.Sprintf("[%s](%s)", creatorName, creatorUrl),
			Inline: true,
		}
		fields = append(fields, &creator)
	}

	return discordgo.MessageEmbed{
		URL:         media.Media.SiteUrl,
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
			fmt.Println("No token provided. Please run: airhorn -t <bot token>")
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
		err = discord.UpdateStatus(0, "A friendly helpful bot!")
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

	if strings.HasPrefix(m.Content, "!anibot") {
		command := strings.TrimPrefix(m.Content, "!anibot ")

		i, err := strconv.Atoi(command)
		if err != nil {
			fmt.Println("Error parsing request", err)
			return
		}
		ctx := context.Background()
		err, res := anilist.MediaFromId(ctx, i)
		if err != nil {
			fmt.Println("Error getting Media", err)
			return
		}

		embed := embed(res)
		s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	}
}
