package tools

import (
	"context"
	"fmt"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/discordbuttons"
	"github.com/bwmarrin/discordgo"
)

type AttachedButton struct {
	Button    discordbuttons.Button
	MessageID string
	ChannelID string
}

// Send an Embed message to the given channel using the provided Session.
func Send(s *discordgo.Session, channel string, media anilist.Media) (err error) {
	embed, err := Embed(media)
	if err != nil {
		return
	}

	sent, err := s.ChannelMessageSendEmbed(channel, &embed)
	if err != nil {
		return
	}

	creator, err := media.Creator()
	fmt.Println(creator)
	if err == nil {
		creatorButton := discordbuttons.Button{
			Data:     nil,
			Reaction: CreatorReaction,
			Callback: func(s *discordgo.Session, r *discordgo.MessageReactionAdd, mID string, cID string, data interface{}) {
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

		discordbuttons.AttachButton(s, sent.ID, sent.ChannelID, creatorButton, true)
		s.MessageReactionAdd(sent.ChannelID, sent.ID, creatorButton.Reaction)
	}
	director, err := media.Director()
	fmt.Println(director)
	if err == nil {
		directorButton := discordbuttons.Button{
			Data:     nil,
			Reaction: DirectorReaction,
			Callback: func(s *discordgo.Session, r *discordgo.MessageReactionAdd, mID string, cID string, data interface{}) {
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
		discordbuttons.AttachButton(s, sent.ID, sent.ChannelID, directorButton, true)
		s.MessageReactionAdd(sent.ChannelID, sent.ID, directorButton.Reaction)
	}
	return
}
