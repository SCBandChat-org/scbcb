package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

//this whole thing is dumb but i wanted it so yeah

func logger(discord *discordgo.Session, m *discordgo.MessageCreate) {
	msg := m.Message
	timestamp := msg.Timestamp
	logToFile(timestamp, msg.ChannelID, msg.Content, msg.Author.ID, msg.Author.Username, msg.Author.Discriminator)
	if msg.Content == "reply" {
		discord.ChannelMessageSend(msg.ChannelID, "replied")
	}
	ch, _ := discord.Channel(msg.ChannelID)
	log.Printf("#%s %s#%s: %s", ch.Name, msg.Author.Username, msg.Author.Discriminator, msg.Content)
}

func logToFile(t time.Time, ch string, con string, a string, un string, disc string) {

	time := t
	fullpath := "./logs/" + fmt.Sprintf("%s/%d/%d/%d/", ch, time.Year(), time.Month(), time.Day())
	path := "./logs/" + fmt.Sprintf("%s/%d/%d/", ch, time.Year(), time.Month())
	if _, err := os.Stat(fullpath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(fullpath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	c := fmt.Sprintf("%d#isBot=%d#%s#%s#%s", time.UnixMilli(), isbot(a), un, disc, con)
	//log to user file
	userLog(path, a, c)

	//log to channel file
	channelLog(fullpath, c)
}

func isbot(a string) int {
	if a == myselfID {
		return 1
	} else {
		return 0
	}
}

func userLog(path string, a string, c string) {
	f, err := os.OpenFile(path+a+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(c + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func channelLog(fullpath string, c string) {
	f, err := os.OpenFile(fullpath+"channel.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(c + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
