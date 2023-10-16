package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix        string
	prefixPattern *regexp.Regexp
)

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       []string
	RoleNeeded  *Role
	Handler     func(caller *discordgo.Member, message *discordgo.Message, args []string) error
}

// List of commands
var Commands = []Command{
	// {
	// 	Name:        "Name",
	// 	Aliases:     []string{"alieas"},
	// 	Description: "self explanitory",
	// 	Usage:       []string{"usage options for help command that doesnt exist"},
	// 	RoleNeeded:  &lowest staff role needed for this command,
	// 	Handler:     handler,
	// },
	{
		Name:        "blankmessage",
		Aliases:     []string{"blank"},
		Description: "Create a blank message",
		Usage:       []string{""},
		RoleNeeded:  &Mod,
		Handler:     blankMessageHandler,
	},
	{
		Name:        "refreshbuttons",
		Aliases:     []string{"rb"},
		Description: "test command",
		Usage:       []string{""},
		Handler:     refreshHandler,
	},
	{
		Name:        "tempmute",
		Aliases:     []string{"tm"},
		Description: "Mute someone temporarily, optionally from a specific channel",
		Usage:       []string{"@user reason", "@user #channel reason", "#channel @user reason"},
		RoleNeeded:  &SCBandChatorg,
		Handler:     muteHandler,
	},
	{
		Name:        "mute",
		Aliases:     []string{"m"},
		Description: "Mute someone permanently, optionally from a specific channel",
		Usage:       []string{"@user reason", "@user #channel reason", "#channel @user reason"},
		RoleNeeded:  &Mod,
		Handler:     muteHandler,
	},
	{
		Name:        "unmute",
		Aliases:     []string{"um"},
		Description: "Unmute someone, either server-wide, from a specific channel, or remove all mutes",
		Usage:       []string{"@user", "@user #channel", "@user all"},
		RoleNeeded:  &Mod,
		Handler:     unmuteHandler,
	},
	{
		Name:        "kick",
		Description: "Kick someone from the server",
		Usage:       []string{"@user reason"},
		RoleNeeded:  &Mod,
		Handler:     rektHandler,
	},
	{
		Name:        "ban",
		Description: "Ban someone from the server",
		Usage:       []string{"@user reason"},
		RoleNeeded:  &Mod,
		Handler:     rektHandler,
	},
}

func initCommands() {
	// Load prefix from the environment
	prefix = os.Getenv("PREFIX")
	log.Println(prefix)
	if prefix == "" {
		prefix = ";"
	}
	// Match case-insensitive & ignore whitespace around prefix
	prefixPattern = regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(prefix) + `\s*`)

	// Have to append helpCommand after initializing Commands to avoid an initialization loop, but i removed help commands for now
	Commands = append(Commands) //, helpCommand)
}

func blankMessageHandler(discord *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ := discord.ChannelMessageSend(m.ChannelID, "** **")
	return
}

func commands(discord *discordgo.Session, m *discordgo.MessageCreate) {
	msg := m.Message

	if msg == nil || msg.Author == nil || msg.Type != discordgo.MessageTypeDefault || msg.Author.ID == myselfID {
		return // wtf
	}
	if msg.GuildID != scbc && msg.GuildID != "" {
		return // Only allow guild messages and DMs
	}

	content := msg.Content
	content = strings.Replace(content, ">", "> ", -1)
	content = strings.Replace(content, "<", " <", -1)

	if match := prefixPattern.FindString(content); match != "" {
		args := strings.Fields(content[len(match):])
		command := findCommand(strings.ToLower(args[0]))
		if command == nil {
			_, _ = discord.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("Command \"%s\" not found! Try %shelp", args[0], prefix))
			return
		}
		author, err := GetMember(msg.Author.ID, msg.GuildID)
		if err != nil {
			return
		}
		if command.RoleNeeded != nil && !IsUserAtLeast(author, *command.RoleNeeded) {
			_ = embed(msg.ChannelID, fmt.Sprintf("Command \"%s\" requires at least %s", command.Name, command.RoleNeeded.Name))
			return
		}
		err = command.Handler(author, msg, args)
		if err != nil {
			_ = errorEmbed(msg.ChannelID, fmt.Sprintf("Command \"%s\" returned an error:\n ```%s```", command.Name, err.Error()))
			return
		}

	}
}
