package main

import (
	"context"
	"fmt"

	"github.com/buckley-w-david/anibot/anilist"
	"github.com/buckley-w-david/discordbuttons"
	"github.com/bwmarrin/discordgo"
)

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
					Send(s, r.ChannelID, media)
				}
			},
		}

		discordbuttons.AttachButton(s, sent.ID, sent.ChannelID, creatorButton, true)
		s.MessageReactionAdd(sent.ChannelID, sent.ID, creatorButton.Reaction)
	}

	director, err := media.Director()
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
					go Send(s, r.ChannelID, media)
				}
			},
		}
		discordbuttons.AttachButton(s, sent.ID, sent.ChannelID, directorButton, true)
		s.MessageReactionAdd(sent.ChannelID, sent.ID, directorButton.Reaction)
	}
	for i := range media.Studios.Edges {
		// If we close over i itself, every callback will be the same, since i is updated in place by the for loop.
		// Be creating a new, locally scoped variable and using that instead, our closure works correctly.
		j := i
		studioButton := discordbuttons.Button{
			Data:     nil,
			Reaction: StudioReactions[i%len(StudioReactions)],
			Callback: func(s *discordgo.Session, r *discordgo.MessageReactionAdd, mID string, cID string, data interface{}) {
				fmt.Println(media.Studios.Edges[j].Studio.ID)
				ctx := context.Background()
				studioMedia, err := anilist.MediaFromStudioID(ctx, media.Studios.Edges[j].Studio.ID, 3)
				if err != nil {
					fmt.Println(err)
					return
				}

				for _, media := range studioMedia {
					go Send(s, r.ChannelID, media)
				}
			},
		}
		discordbuttons.AttachButton(s, sent.ID, sent.ChannelID, studioButton, true)
		s.MessageReactionAdd(sent.ChannelID, sent.ID, studioButton.Reaction)
	}
	return
}
