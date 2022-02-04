package main

import (
	"log"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
)

func embed(ch string, text string) error {
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       prettyembedcolor,
		Description: text,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "SCBandChat.org",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err := discord.ChannelMessageSendEmbed(ch, embed)
	return err
}

func errorEmbed(ch string, text string) error {
	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       errorColor,
		Description: text,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "SCBandChat.org",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err := discord.ChannelMessageSendEmbed(ch, embed)
	return err
}

func GetMember(userID string, guildID string) (member *discordgo.Member, err error) {
	member, err = discord.State.Member(guildID, userID)
	if err != nil {
		log.Println(err)
		member, err = discord.GuildMember(guildID, userID)
		log.Println(err)
	}
	return member, err
}

func findNamedMatches(r *regexp.Regexp, str string) map[string]string {
	matches := r.FindStringSubmatch(str)
	names := r.SubexpNames()
	subs := map[string]string{}

	for i, sub := range matches {
		if names[i] != "" {
			subs[names[i]] = sub
		}
	}

	return subs
}

//this is a shitty way to get a random entry, but it works
func getFromMap(m map[int]string) (int, string) {
	for k, v := range m {
		return k, v
	}
	return 0, "no status"
}

func statusUpdate(splashText map[int]string) error {

	activityType, activityText := getFromMap(splashText)
	log.Println(activityType, activityText)

	err := discord.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince: nil,
		Activities: []*discordgo.Activity{{
			Name: activityText,
			Type: discordgo.ActivityType(activityType),
		}},
		AFK:    false,
		Status: "",
	})
	return err
}

func hasRole(user *discordgo.Member, role ...Role) bool {
	for _, r := range role {
		if includes(user.Roles, r.ID) {
			return true
		}
	}
	return false
}

// True if user has ALL roles passed in
func hasRoles(user *discordgo.Member, role ...Role) bool {
	for _, r := range role {
		if !includes(user.Roles, r.ID) {
			return false
		}
	}
	return true
}

func includes(list []string, val string) bool {
	for _, x := range list {
		if x == val {
			return true
		}
	}
	return false
}

func findCommand(command string) *Command {
	for _, it := range Commands {
		if command == it.Name {
			return &it
		}
		for _, alias := range it.Aliases {
			if command == alias {
				return &it
			}
		}
	}
	return nil
}

func outranks(user1, user2 *discordgo.Member) bool {
	role := highestRole(user2)
	if role == nil {
		return IsUserStaff(user1)
	}
	return IsUserHigherThan(user1, *role)
}

func highestRole(user *discordgo.Member) *Role {
	for _, role := range staffRoles {
		if includes(user.Roles, role.ID) {
			return &role
		}
	}
	return nil
}
