package shared

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/bwmarrin/discordgo"
)

type CliOption struct {
	name         string
	short        string
	defaultValue string
	value        string
	description  string
}

func (cmd *CliOption) StringVar() {
	flag.StringVar(&cmd.value, cmd.short, os.Getenv(strings.ToUpper(cmd.name)), cmd.description)
}

func (cmd CliOption) OrEnv() (value string, err error) {
	if cmd.value != "" {
		value = cmd.value
	} else {
		if cmd.defaultValue != "" {
			value = cmd.defaultValue
		} else {
			err = errors.New("Unable to find value")
		}
	}
	return
}

var (
	Token CliOption
	Hook  CliOption

	hookURL      string
	MissingToken string
)

func init() {
	Token = CliOption{name: "token", short: "t", description: "Bot Token"}
	Hook = CliOption{name: "hook", short: "hook", description: "Webook Server URL"}

	Token.StringVar()
	Hook.StringVar()
	flag.Parse()

	MissingToken = "No token provided. Please run: anibot -t <bot token>"
}

func embed(media anilist.MediaResponse, channel string, hookURL string) (discordgo.MessageEmbed, error) {
	hooks := channel != "" && hookURL != ""
	var URL *url.URL
	if hooks {
		var err error
		URL, err = url.Parse(hookURL)
		if err != nil {
			fmt.Println(hookURL)
			return discordgo.MessageEmbed{}, errors.New("Bad hookURL")
		}
	}

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
		var value string
		if hooks {
			parameters := url.Values{}
			parameters.Add("channel", channel)
			parameters.Add("type", "studio")
			parameters.Add("input", strconv.Itoa(studio.Studio.ID))
			URL.RawQuery = parameters.Encode()

			value = fmt.Sprintf(
				"[%s](%s) - [Preview](%s)",
				studio.Studio.Name,
				studio.Studio.SiteURL,
				URL.String(),
			)
		} else {
			value = fmt.Sprintf("[%s](%s)", studio.Studio.Name, studio.Studio.SiteURL)
		}
		studios[i] = &discordgo.MessageEmbedField{
			Name:   "Studio",
			Value:  value,
			Inline: false,
		}
	}
	fields = append(fields, studios...)

	director, err := media.Director()

	if err == nil {
		var value string
		if hooks {
			parameters := url.Values{}
			parameters.Add("channel", channel)
			parameters.Add("type", "person")
			parameters.Add("input", strconv.Itoa(director.ID))
			URL.RawQuery = parameters.Encode()

			value = fmt.Sprintf(
				"[%s %s](%s) - [Preview](%s)",
				director.Name.First,
				director.Name.Last,
				director.SiteURL,
				URL.String(),
			)
		} else {
			value = fmt.Sprintf("[%s %s](%s)", director.Name.First, director.Name.Last, director.SiteURL)
		}
		director := discordgo.MessageEmbedField{
			Name:   "Director",
			Value:  value,
			Inline: true,
		}
		fields = append(fields, &director)
	}

	creator, err := media.Creator()

	if err == nil {
		var value string
		if hooks {
			parameters := url.Values{}
			parameters.Add("channel", channel)
			parameters.Add("type", "person")
			parameters.Add("input", strconv.Itoa(creator.ID))
			URL.RawQuery = parameters.Encode()

			value = fmt.Sprintf(
				"[%s %s](%s) - [Preview](%s)",
				creator.Name.First,
				creator.Name.Last,
				creator.SiteURL,
				URL.String(),
			)
		} else {
			value = fmt.Sprintf("[%s %s](%s)", creator.Name.First, creator.Name.Last, creator.SiteURL)
		}

		creator := discordgo.MessageEmbedField{
			Name:   "Original Creator",
			Value:  value,
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
	}, nil
}

func Embed(media anilist.MediaResponse) (discordgo.MessageEmbed, error) {
	return embed(media, "", "")
}

func EmbedHook(media anilist.MediaResponse, channel string, hookURL string) (discordgo.MessageEmbed, error) {
	return embed(media, channel, hookURL)
}
