package discord

import (
	. "backend/config"
	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session
var me *discordgo.User

func Init() error {
	var err error

	if err := explorerInit(); err != nil {
		return err
	}

	if session, err = discordgo.New("Bot " + Config.DiscordToken); err != nil {
		return err
	}

	if me, err = session.User("@me"); err != nil {
		return err
	}

	if err := register(); err != nil {
		return err
	}

	if err := session.Open(); err != nil {
		return err
	}

	go tick()

	return nil
}

func register() error {
	commands := map[string]struct {
		handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
		command *discordgo.ApplicationCommand
	}{
		"latest": {
			handler: latestSales,
			command: &discordgo.ApplicationCommand{
				Description: "List latest sales of our NFT",
			},
		},
	}

	for name, h := range commands {
		h.command.Name = name

		_, err := session.ApplicationCommandCreate(me.ID, "", h.command)
		if err != nil {
			return err
		}
	}

	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.Type == discordgo.InteractionMessageComponent {
			data := interaction.MessageComponentData()
			if h, ok := commands[data.CustomID]; ok {
				h.handler(session, interaction)
				return
			}
			return
		}

		data := interaction.ApplicationCommandData()
		if h, ok := commands[data.Name]; ok {
			h.handler(session, interaction)
		}
	})

	return nil
}

func Close() error {
	return session.Close()
}
