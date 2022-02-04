package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/subosito/gotenv"
)

var discord *discordgo.Session

var myselfID string

func init() {
	var err error

	// You can set environment variables in the git-ignored .env file for convenience while running locally
	err = gotenv.Load()
	if err == nil {
		log.Println("Loaded .env file")
	} else if os.IsNotExist(err) {
		log.Println("No .env file found")
		err = nil // Mutating state is bad mkay
	} else {
		panic(err)
	}

	token := os.Getenv("DISCORD_BOT_TOKEN") // DISCORD_BOT_TOKEN || dev_token
	if token == "" {
		panic("Must set environment variable DISCORD_BOT_TOKEN")
	}
	log.Println("Establishing discord connection")
	discord, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences | discordgo.IntentsAllWithoutPrivileged)
	user, err := discord.User("@me")
	if err != nil {
		panic(err)
	}

	myselfID = user.ID
	log.Printf("I am %s, %s#%s", user.ID, user.Username, user.Discriminator)

	discord.AddHandler(onReady)
	discord.AddHandler(logger)
	discord.AddHandler(commands)
	discord.AddHandler(roleButtonsListener)
	discord.AddHandler(remute)

}

func main() {
	err := discord.Open()
	if err != nil {
		panic(err)
	}
	log.Println("Connected to discord")
	forever := make(chan int)
	<-forever
}

func onReady(discord *discordgo.Session, ready *discordgo.Ready) {
	//roleButtons()
	initCommands()

	err := statusUpdate(splashText)
	if err != nil {
		log.Println("Error attempting to set my status")
		log.Println(err)
	}

	servers := discord.State.Guilds
	log.Printf("bot has started on %d servers:", len(servers))
	for _, guild := range servers {
		log.Println("Server ID", guild.ID)
		fullGuild, err := discord.Guild(guild.ID)
		if err == nil {
			log.Println("Full server ID", fullGuild.ID, "Full name", fullGuild.Name)
		}
	}
}
