package tools

import (
	"context"
	"fmt"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/discordbuttons"
	"github.com/bwmarrin/discordgo"
)

// Send an Embed message to the given channel using the provided Session.
func Send(s *discordgo.Session, channel string, media anilist.Media) error {
	embed, err := Embed(media)
	if err != nil {
		return err
	}

	sent, err := s.ChannelMessageSendEmbed(channel, &embed)
	if err != nil {
		return err
	}

	creator, err := media.Creator()
	fmt.Println(creator)
	if err == nil {
		creatorButton := discordbuttons.Button{
			Data:     nil,
			Reaction: CreatorReaction,
			Callback: func(s *discordgo.Session, r *discordgo.MessageReactionAdd, m *discordgo.Message, data interface{}) {
				fmt.Println(creator.ID)
				ctx := context.Background()
				creatorMedia, err := anilist.MediaFromPersonID(ctx, creator.ID, 3)
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, media := range creatorMedia {
					fmt.Println(media.SiteURL)
					Send(s, r.ChannelID, media)
				}
			},
		}

		discordbuttons.AddButton(s, sent, creatorButton, true)
	}
	director, err := media.Director()
	fmt.Println(director)
	if err == nil {
		directorButton := discordbuttons.Button{
			Data:     nil,
			Reaction: DirectorReaction,
			Callback: func(s *discordgo.Session, r *discordgo.MessageReactionAdd, m *discordgo.Message, data interface{}) {
				fmt.Println(director.ID)
				ctx := context.Background()
				directorMedia, err := anilist.MediaFromPersonID(ctx, director.ID, 3)
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, media := range directorMedia {
					fmt.Println(media.SiteURL)
					Send(s, r.ChannelID, media)
				}
			},
		}
		discordbuttons.AddButton(s, sent, directorButton, true)
	}
	return nil
}
