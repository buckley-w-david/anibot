package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/machinebox/graphql"
)

var (
	token  string
	client *graphql.Client
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
	client = graphql.NewClient("https://graphql.anilist.co/")
}

// {"Media":{"siteUrl":"https:\/\/anilist.co\/anime\/2547"}}
type MediaResponse struct {
	Media struct {
		SiteUrl string `json:"siteUrl"`
		Title   struct {
			English string `json:"english,omitempty"`
			Romaji  string `json:"romaji,omitempty"`
		} `json:"title"`
		Description string `json:"description"`
		CoverImage  struct {
			ExtraLarge string `json:"extraLarge,omitempty"`
			Large      string `json:"large,omitempty"`
			Medium     string `json:"medium,omitempty"`
		} `json:"coverImage"`
		MediaType string `json:"type"`
		Format    string `json:"format"`
		Source    string `json:"source"`
		Studios   struct {
			Edges []struct {
				Node struct {
					Name    string `json:"name"`
					SiteUrl string `json:"siteUrl"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"studios"`
		Staff struct {
			Edges []struct {
				Role string `json:"role"`
				Node struct {
					SiteUrl string `json:"siteUrl"`
					Name    struct {
						First string `json:"first"`
						Last  string `json:"last"`
					} `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"staff"`
	} `json:"Media"`
}

func (media MediaResponse) director() (error, string, string) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Director" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return nil, name, media.Media.Staff.Edges[i].Node.SiteUrl
		}
	}
	return errors.New("Unable to find director"), "", ""
}

func (media MediaResponse) creator() (error, string, string) {
	for i := range media.Media.Staff.Edges {
		if media.Media.Staff.Edges[i].Role == "Original Creator" {
			name := fmt.Sprintf("%s %s", media.Media.Staff.Edges[i].Node.Name.First, media.Media.Staff.Edges[i].Node.Name.Last)
			return nil, name, media.Media.Staff.Edges[i].Node.SiteUrl
		}
	}
	return errors.New("Unable to find director"), "", ""
}

func (media MediaResponse) embed() discordgo.MessageEmbed {
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

	directorErr, directorName, directorUrl := media.director()

	var director discordgo.MessageEmbedField
	if directorErr == nil {
		director = discordgo.MessageEmbedField{
			Name:   "Director",
			Value:  fmt.Sprintf("[%s](%s)", directorName, directorUrl),
			Inline: true,
		}
		fields = append(fields, &director)
	}

	creatorErr, creatorName, creatorUrl := media.creator()

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
		fmt.Println("No token provided. Please run: airhorn -t <bot token>")
		return
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

		req := graphql.NewRequest(`
		  query ($id: Int!) {
            Media(id: $id) {
              siteUrl
              title{
                english
                romaji
              }
              description(asHtml: false)
              coverImage {
		  	  extraLarge
		  	  large
                medium
              }
              type
              format
              source
              studios {
                edges {
                  node {
                    name
                    siteUrl
                  }
                }
              }
              staff {
                edges {
                  role
                  node{
                    siteUrl
                    name{
                      first
                      last
                    }
                  }
                }
              }
            }
		  }
	    `)
		i, err := strconv.Atoi(command)
		if err != nil {
			fmt.Println("Error parsing request", err)
			return
		}
		req.Var("id", i)
		ctx := context.Background()

		var res MediaResponse
		if err := client.Run(ctx, req, &res); err != nil {
			fmt.Println("Error making GraphQL query", err)
			return
		}
		spew.Dump(res)

		embed := res.embed()
		s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	}
}
