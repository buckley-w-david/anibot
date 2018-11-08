package shared

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/bwmarrin/discordgo"
)

type CliOption struct {
	Name         string
	Short        string
	DefaultValue string
	Value        string
	Description  string
}

func (cmd *CliOption) StringVar() {
	flag.StringVar(&cmd.Value, cmd.Short, os.Getenv(strings.ToUpper(cmd.Name)), cmd.Description)
}

func (cmd CliOption) OrEnv() (value string, err error) {
	if cmd.Value != "" {
		value = cmd.Value
	} else {
		if cmd.DefaultValue != "" {
			value = cmd.DefaultValue
		} else {
			err = errors.New("Unable to find value")
		}
	}
	return
}

var (
	Token CliOption

	MissingToken     string
	DirectorReaction string
	CreatorReaction  string
	StudioReactions  []string
)

func init() {
	MissingToken = "No token provided. Please run: anibot -t <bot token>"

	DirectorReaction = "👉"
	CreatorReaction = "👈"
	StudioReactions = []string{"👇", "👆"}
}

func SetupSharedOptions() {
	Token = CliOption{Name: "token", Short: "t", Description: "Bot Token"}

	Token.StringVar()
}

func Emojis() []string {
	emojis := []string{DirectorReaction, CreatorReaction}
	emojis = append(emojis, StudioReactions...)
	return emojis
}

func Reaction(m discordgo.Message, reaction string) (discordgo.MessageEmbedField, error) {
	for _, embed := range m.Embeds {
		for _, field := range embed.Fields {
			if strings.HasSuffix(field.Name, reaction) {
				return *field, nil
			}
		}
	}
	return discordgo.MessageEmbedField{}, errors.New("Reaction not present in Fields")
}

func Embed(media anilist.MediaResponse) (discordgo.MessageEmbed, error) {
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
		value := fmt.Sprintf("[%s](%s)", studio.Studio.Name, studio.Studio.SiteURL)
		studios[i] = &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Studio %s", StudioReactions[i%len(StudioReactions)]),
			Value:  value,
			Inline: false,
		}
	}
	fields = append(fields, studios...)

	director, err := media.Director()
	if err == nil {
		value := fmt.Sprintf("[%s %s](%s)", director.Name.First, director.Name.Last, director.SiteURL)
		director := discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Director %s", DirectorReaction),
			Value:  value,
			Inline: true,
		}
		fields = append(fields, &director)
	}

	creator, err := media.Creator()
	if err == nil {
		value := fmt.Sprintf("[%s %s](%s)", creator.Name.First, creator.Name.Last, creator.SiteURL)

		creator := discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Original Creator %s", CreatorReaction),
			Value:  value,
			Inline: true,
		}
		fields = append(fields, &creator)
	}

	return discordgo.MessageEmbed{
		URL:         media.Media.SiteURL,
		Title:       media.Media.Title.Romaji,
		Description: strings.Replace(media.Media.Description, "<br>", "\n", -1),
		Color:       0x00ff00,
		Thumbnail:   &coverImage,
		Fields:      fields,
	}, nil
}
