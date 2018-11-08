package tools

import (
	"github.com/buckley-w-david/anibot/anilist"
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
	for _, emoji := range Emojis() {
		s.MessageReactionAdd(sent.ChannelID, sent.ID, emoji)
		if err != nil {
			return err
		}
	}

	return nil
}
