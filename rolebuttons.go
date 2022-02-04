package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func roleButtons() {

	channelID := "933914370952532009"

	messageID1 := "933919451898839040"
	messageID2 := "933919465702322307"
	messageID3 := "933919473767956491"
	messageID4 := "933919488380928000"
	messageID5 := "933935681045155874"
	messageID6 := "933935711596466188"
	messageID7 := "933937206425428008"

	//this is the worst way to do this, but it stops rare errors
	time.Sleep(1 * time.Second)

	//this is also awful but i dont care
	updateMessage(schools1, messageID1, channelID)
	updateMessage(schools2, messageID2, channelID)
	updateMessage(classes, messageID3, channelID)
	updateMessage(schoolSize, messageID4, channelID)
	updateMessage(instruments1, messageID5, channelID)
	updateMessage(instruments2, messageID6, channelID)
	updateMessage(leadership, messageID7, channelID)
}
func updateMessage(a []string, m string, c string) {
	com := getButtonsForList(a)
	_, err := discord.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         m,
		Channel:    c,
		Components: com,
	})
	if err != nil {
		log.Println(err)
		return
	}
}
func roleButtonsListener(discord *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Interaction.ChannelID != "933914370952532009" {
		return
	}
	c := i.MessageComponentData().CustomID
	m := i.Member
	r, _ := discord.State.Role(i.GuildID, c)
	if containsRole(m.Roles, r.ID) {
		err := discord.GuildMemberRoleRemove(i.GuildID, m.User.ID, r.ID)
		if err != nil {
			log.Println(err)
			return
		}
		discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Removed Role: " + r.Name, Flags: 1 << 6}})
	} else {
		err := discord.GuildMemberRoleAdd(i.GuildID, m.User.ID, r.ID)
		if err != nil {
			log.Println(err)
			return
		}
		discord.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Applied Role: " + r.Name, Flags: 1 << 6}})
	}
}

func containsRole(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getButtonsForList(a []string) []discordgo.MessageComponent {
	if len(a) > 25 {
		log.Println("too many roles")
		return nil
	}
	n := make([]string, 0)
	var com []discordgo.MessageComponent
	for _, id := range a {
		ro, _ := discord.State.Role(scbc, id)
		n = append(n, ro.Name)
	}
	aa := chunkSlice(a, 5)
	nn := chunkSlice(n, 5)
	for i, ch := range nn {
		var ar discordgo.ActionsRow
		for u, ro := range ch {
			var bt discordgo.Button
			log.Printf("Loaded button: %s", ro)
			bt.Label = nn[i][u]
			bt.CustomID = aa[i][u]
			bt.Style = discordgo.PrimaryButton
			ar.Components = append(ar.Components, bt)
		}
		com = append(com, ar)
	}
	return com
}

// []discordgo.MessageComponent{
// 	discordgo.ActionsRow{
// 		Components: []discordgo.MessageComponent{
// 			discordgo.Button{
// 				Label:    "I'm going",
// 				Style:    discordgo.SuccessButton,
// 				CustomID: "going",
// 			},
// 			discordgo.Button{
// 				Label:    "I'm FLAKING",
// 				Style:    discordgo.DangerButton,
// 				CustomID: "flaking",
// 			},
// 		},
// 	},
// },

// func getName(s string) chan string {
// 	r := make(chan string)
// 	log.Println("made r")
// 	ro, _ := discord.State.Role("889590337700495381", s)
// 	log.Println(ro.Name)
// 	r <- ro.Name
// 	log.Println("sent ro")
// 	return r
// }

func chunkSlice(items []string, chunkSize int32) (chunks [][]string) {
	//While there are more items remaining than chunkSize...
	for chunkSize < int32(len(items)) {
		//We take a slice of size chunkSize from the items array and append it to the new array
		chunks = append(chunks, items[0:chunkSize])
		//Then we remove those elements from the items array
		items = items[chunkSize:]
	}
	//Finally we append the remaining items to the new array and return it
	return append(chunks, items)
}

func DerefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

//remove this
func refreshHandler(caller *discordgo.Member, msg *discordgo.Message, args []string) error {
	pendingEmbed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       prettyembedcolor,
		Description: "refreshing buttons, this might take a moment...",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "SCBandChat.org",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	pending, err := discord.ChannelMessageSendEmbed(msg.ChannelID, pendingEmbed)
	if err != nil {
		return err
	}

	roleButtons()
	complete := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       prettyembedcolor,
		Description: "buttons refreshed",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "SCBandChat.org",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = discord.ChannelMessageEditEmbed(pending.ChannelID, pending.ID, complete)
	return err
}
