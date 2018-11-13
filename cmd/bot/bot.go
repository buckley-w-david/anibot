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

	"github.com/buckley-w-david/anibot/pkg/anilist"
	"github.com/bwmarrin/discordgo"
)

var (
	discord *discordgo.Session

	re *regexp.Regexp
)

func init() {
	SetupSharedOptions()
	flag.Parse()

	re = regexp.MustCompile(`{(?P<ANIME>.*?)}|<(?P<MANGA>.*?)>`)
}

func main() {
	botToken, err := Token.OrEnv()
	if err != nil {
		fmt.Println(MissingToken)
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
	// discord.AddHandler(reaction)

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

// func reaction(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
// 	fmt.Println(m.Emoji)
// 	spew.Dump(m.Emoji)
// }

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
	matches := re.FindAllStringSubmatch(m.Content, -1)
	for i := range matches {
		j := i
		// TODO: This doesn't really need to be a local function, factor out into a real one
		go func() {
			var requestType anilist.MediaType
			var mediaName string
			match := matches[j]
			if match[1] != "" {
				requestType = anilist.ANIME
				mediaName = match[1]
			} else if match[2] != "" {
				requestType = anilist.MANGA
				mediaName = match[2]
			} else {
				fmt.Println("Weird match: ", match)
				return
			}

			// See if we can get an exact title match
			query := anilist.MediaQuery{
				Title:      mediaName,
				Type:       requestType.String(),
				Sort:       []string{"SEARCH_MATCH"},
				MaxResults: 3,
			}
			potentials, err := anilist.MediaFromMediaQuery(ctx, query)
			if err != nil {
				fmt.Println("Error getting Media", err)
				return
			}
			for _, potential := range potentials {
				if potential.Title.Romaji == mediaName || potential.Title.English == mediaName {
					go Send(s, m.ChannelID, potential)
					return
				}
			}

			// If not just grab the most popular one
			query = anilist.MediaQuery{
				Title:      mediaName,
				Type:       requestType.String(),
				Sort:       []string{"POPULARITY_DESC"},
				MaxResults: 1,
			}
			media, err := anilist.MediaFromMediaQuery(ctx, query)
			if err != nil {
				fmt.Println("Error getting Media", err)
				return
			}

			if len(media) > 0 {
				go Send(s, m.ChannelID, media[0])
			}
		}()
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
			Send(s, channel, media)
		}
	case "person":
		for _, name := range args {
			ctx := context.Background()
			medias, err := anilist.MediaFromPersonQuery(ctx, anilist.PersonQuery{Name: name, Type: queryType, MaxResults: 1})
			if err != nil {
				return err
			}
			media = medias[0]
			Send(s, channel, media)
		}
	case "studio":
		for _, name := range args {
			ctx := context.Background()
			medias, err := anilist.MediaFromStudioQuery(ctx, anilist.StudioQuery{Name: name, MaxResults: 1})
			if err != nil {
				return err
			}
			media = medias[0]
			Send(s, channel, media)
		}
	default:
		fmt.Println("default")
	}
	return nil
}
