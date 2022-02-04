package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const (
	RateLimit        = 5 * time.Minute
	TempmuteDuration = 3 * time.Hour
	UnmuteInterval   = 10 * time.Second
)

var ratelimit = make(map[string]int64)
var ratelimitLock sync.Mutex

var mentionRegex = regexp.MustCompile(`^<(?P<Type>(@|#|@&))!?(?P<ID>\d+)>$`)

func evalRatelimit(author string) bool {
	ratelimitLock.Lock()
	defer ratelimitLock.Unlock()

	until := ratelimit[author]
	if until < time.Now().UnixNano() { // defaults to 0 so this works properly
		rateLimitDuration := RateLimit
		ratelimit[author] = time.Now().Add(rateLimitDuration).UnixNano()
		return true
	}

	return false
}

// Turns the first one or two args into users and/or channels and also returns whatever args weren't consumed
func getUserAndChannelAndArgs(args []string) (user *discordgo.User, channel *discordgo.Channel, remainingArgs []string) {
	remainingArgs = args
	if len(args) < 1 {
		return
	}
	user, channel = getUserOrChannelForArg(args[0])
	if user != nil || channel != nil {
		// Consume an arg
		remainingArgs = remainingArgs[1:]
		// Don't return since we want to try the second arg too
	} else {
		// No match on first arg so don't try to match second arg
		return
	}

	// getUserOrChannelForArg always has one nil arg, so if-else instead of if-elseif is fine
	if len(args) < 2 {
		return
	}
	if user == nil {
		user, _ = getUserOrChannelForArg(args[1])
		if user != nil {
			// Consume an arg
			remainingArgs = remainingArgs[1:]
		}
	} else {
		_, channel = getUserOrChannelForArg(args[1])
		if channel != nil {
			// Consume an arg
			remainingArgs = remainingArgs[1:]
		}
	}
	return
}

// Send a blocking api request if a match is found
func getUserOrChannelForArg(arg string) (*discordgo.User, *discordgo.Channel) {
	match := findNamedMatches(mentionRegex, arg)
	id := match["ID"]
	if id == "" {
		return nil, nil
	}

	// Sends a blocking API request
	switch match["Type"] {
	case "#":
		{
			// Try to get from the "State" cache before falling back to the API
			channel, err := discord.State.Channel(id)
			if err != nil || channel == nil {
				channel, err = discord.Channel(id)
			}
			if err == nil {
				return nil, channel
			}
		}
	case "@":
		{
			// discordgo doesn't cache users, only guilds, channels & members
			user, err := discord.User(id)
			if err == nil {
				return user, nil
			}
		}
	}

	return nil, nil
}

// Gets the mute role associated with the given channel
// Returns an error if no matching role exists
func getMuteRoleForChannel(channel *discordgo.Channel) (role string, err error) {
	if channel == nil {
		role = muteRoles[""]
		return
	}
	var ok bool // Avoid shadowing role with :=
	if role, ok = muteRoles[channel.ID]; !ok {
		err = fmt.Errorf("unable to find mute role for channel %s", channel.Mention())
	}
	return
}

func muteHandler(caller *discordgo.Member, msg *discordgo.Message, args []string) error {
	user, channel, remainingArgs := getUserAndChannelAndArgs(args[1:])
	if user == nil {
		return errors.New("First argument should mention user or channel")
	}
	if len(remainingArgs) < 1 {
		return errors.New("Give a reason")
	}

	target, err := GetMember(user.ID, scbc)
	if err != nil {
		return err
	}

	if !outranks(caller, target) {
		return fmt.Errorf("You don't outrank %s", target.User.Username)
	}

	// // non mods can tempmute, but only on users without roles
	// if strings.ToLower(args[0]) == "tempmute" {

	// 	if IsUserLowerThan(caller, Mod) && !evalRatelimit(msg.Author.ID) {
	// 		return errors.New("Too soon")
	// 	}
	// 	if IsUserLowerThan(caller, Mod) {
	// 		trustedRoles := append(RolesToIDs(staffRoles)) // TODO calculate this only once?
	// 		for _, role := range target.Roles {
	// 			if includes(trustedRoles, role) {
	// 				return errors.New("They have trusted role(s)")
	// 			}
	// 		}
	// 	}
	// }

	muteRole, err := getMuteRoleForChannel(channel)
	if err != nil {
		return fmt.Errorf("Can't mute from %s yet", channel.Mention())
	}

	// Reasons are important
	var reason strings.Builder
	reason.WriteString(args[0])
	reason.WriteString(" has been issued to " + user.Username)
	if channel != nil {
		reason.WriteString(" from channel " + channel.Mention())
	}
	reason.WriteString(" by @" + msg.Author.Username + "#" + msg.Author.Discriminator)
	reason.WriteString(" for reason: " + strings.Join(remainingArgs, " "))
	providedReason := reason.String()

	// Direct message the user being muted
	DM, err := discord.UserChannelCreate(user.ID) // only creates it if it doesn"t already exist
	if err == nil {
		// if there is an error DMing them, we still want to ban them, they just won't know why
		err = embed(DM.ID, providedReason)
		if err != nil {
			log.Printf("Error direct messaging %s#%s: %s\n", user.Username, user.Discriminator, err.Error())
		}
	}

	if db == nil {
		return errors.New("I have no database, so I cannot ")
	}

	// Row values
	var (
		id        uuid.UUID
		userId    = user.ID
		channelId sql.NullString
	)
	if channel != nil {
		channelId = sql.NullString{
			String: channel.ID,
			Valid:  true,
		}
	}
	var expiration *time.Time
	if strings.ToLower(args[0]) == "tempmute" {
		exp := time.Now().Add(TempmuteDuration) // go doesn't let you do this in one line
		expiration = &exp
	}

	// Check if we have a matching mute already
	err = db.QueryRow("SELECT id from mutes WHERE discord_id=$1 AND (channel_id=$2 OR ($2 IS NULL AND channel_id IS NULL))", userId, channelId).Scan(&id)

	if err == nil {
		// Update existing entry, but only if it was already a tempmute. Don't downgrade a mute to a tempmute.
		_, err = db.Exec("UPDATE mutes SET expiration=$2 WHERE id=$1 AND expiration IS NOT NULL", id, expiration)
		if err != nil {
			return err
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		// Insert new entry
		_, err = db.Exec("INSERT INTO mutes (id, discord_id, channel_id, expiration) VALUES ($1, $2, $3, $4)", uuid.New().String(), userId, channelId, expiration)
		if err != nil {
			return err
		}
	} else {
		// Unknown error
		return err
	}
	err = discord.GuildMemberRoleAdd(msg.GuildID, user.ID, muteRole)
	if err != nil {
		return err
	}

	_ = embed(modLog, providedReason)

	_ = embed(msg.ChannelID, providedReason)
	return nil
}

func unmuteHandler(caller *discordgo.Member, msg *discordgo.Message, args []string) (err error) {
	var unmuteAll bool
	user, channel, remainingArgs := getUserAndChannelAndArgs(args[1:])
	if user == nil {
		return errors.New("First argument should mention user")
	}
	if channel == nil && len(remainingArgs) == 1 && "all" == strings.ToLower(remainingArgs[0]) {
		unmuteAll = true
	} else if len(remainingArgs) > 0 {
		return fmt.Errorf("Unexpected arguments \"%s\"", strings.Join(remainingArgs, " "))
	}

	var target *discordgo.Member
	target, err = GetMember(user.ID, scbc)
	if err != nil {
		return
	}

	if outranks(target, caller) {
		return fmt.Errorf("You must be at least the same rank as %s", target.User.Username)
	}

	// produce a list of muted channels for the command output
	var fullMute bool
	var channels []string

	if unmuteAll {
		// Remove all mutes from db
		_, err = db.Exec("DELETE FROM mutes WHERE discord_id = $1", user.ID)
		if err != nil {
			return err
		}

		// Remove all mute roles
		var count uint
		for mutedChannel, muteRole := range muteRoles {
			if hasRole(target, Role{ID: muteRole}) {
				// Keep track of what was unmuted for the reply
				count++
				if mutedChannel == "" {
					fullMute = true
				} else {
					channels = append(channels, mutedChannel)
				}

				// Remove the role
				err = discord.GuildMemberRoleRemove(msg.GuildID, user.ID, muteRole)
			}
		}

		// We didn't actually unmute anything!
		if count < 1 {
			return fmt.Errorf("No mutes found for @%s#%s", user.Username, user.Discriminator)
		}
	} else {
		// unmute specified channel (or server-wide for nil)
		muteRole, err := getMuteRoleForChannel(channel)
		if err != nil {
			if channel != nil {
				return fmt.Errorf("Can't unmute from %s yet", channel.Mention())
			} else {
				return err // Unknown error
			}
		}

		// Check the user is actually muted...
		if !hasRole(target, Role{ID: muteRole}) {
			if channel == nil {
				return fmt.Errorf("%s isn't muted serverwide", user.Username)
			} else {
				return fmt.Errorf("%s isn't muted in %s", user.Username, channel.Mention())
			}
		}

		// Update the database and keep track of what was unmuted (for the reply)
		if channel == nil {
			fullMute = true
			_, err = db.Exec("DELETE FROM mutes WHERE discord_id = $1 AND channel_id IS NULL", user.ID)
			if err != nil {
				return err
			}
		} else {
			channels = append(channels, channel.ID)
			_, err = db.Exec("DELETE FROM mutes WHERE discord_id = $1 AND channel_id = $2", user.ID, channel.ID)
			if err != nil {
				return err
			}
		}

		// Unmute them
		err = discord.GuildMemberRoleRemove(msg.GuildID, user.ID, muteRole)
	}

	// Construct a reply out of the unmuted channels slice
	var reply strings.Builder
	reply.WriteString("User " + user.Username + " has been ")
	if fullMute {
		reply.WriteString("unmuted serverwide")
		if len(channels) > 0 {
			reply.WriteString(" and ")
		}
	}
	if len(channels) > 0 {
		reply.WriteString("unmuted from ")
		for index, id := range channels {
			if index > 0 {
				reply.WriteString(", ")
				if index == len(channels)-1 {
					reply.WriteString("& ")
				}
			}
			reply.WriteString(fmt.Sprintf("<#%s>", id))
		}
	}
	reply.WriteString(" by @" + msg.Author.Username + "#" + msg.Author.Discriminator)
	reply.WriteString("\n")

	// Direct message the user being unmuted
	DM, err := discord.UserChannelCreate(user.ID)
	if err == nil {
		err = embed(DM.ID, reply.String())
		if err != nil {
			s := fmt.Sprintf("Error direct messaging %s#%s: %s", user.Username, user.Discriminator, err.Error())
			log.Println(s)
			reply.WriteString(s + "\n")
		}
	}

	_ = embed(modLog, reply.String())
	_ = embed(msg.ChannelID, reply.String())

	return nil
}

// tbh should this be separate handlers??
// or maybe multiple handlers here is stupid, this is mostly a copy of mute handler :\
func rektHandler(caller *discordgo.Member, msg *discordgo.Message, args []string) error {
	user, channel, remainingArgs := getUserAndChannelAndArgs(args[1:])
	if user == nil {
		return errors.New("First argument should mention user")
	}
	if channel != nil {
		return errors.New(args[0] + " does not support channel mentions")
	}
	if len(remainingArgs) < 1 {
		return errors.New("Give a reason")
	}

	target, err := GetMember(user.ID, scbc)
	if err != nil {
		return err
	}
	if !outranks(caller, target) {
		return fmt.Errorf("You don't outrank %s", target.User.Username)
	}

	// Reasons are important
	providedReason := args[0] + " has been issued to " + user.Username + " by @" + msg.Author.Username + "#" + msg.Author.Discriminator + " for reason: " + strings.Join(remainingArgs, " ")

	// Direct message the user being rekt
	DM, err := discord.UserChannelCreate(user.ID) // only creates it if it doesn"t already exist
	if err == nil {
		// if there is an error DMing them, we still want to ban them, they just won't know why
		err = embed(DM.ID, providedReason)
		if err != nil {
			log.Printf("Error direct messaging %s#%s: %s\n", user.Username, user.Discriminator, err.Error())
		}
	}

	switch args[0] {
	case "ban":
		err = discord.GuildBanCreateWithReason(msg.GuildID, user.ID, providedReason, 0)
	case "kick":
		err = discord.GuildMemberDeleteWithReason(msg.GuildID, user.ID, providedReason)
	}

	if err != nil {
		return err
	}

	_ = embed(modLog, providedReason)

	_ = embed(msg.ChannelID, providedReason)
	return nil
}

// unmuteCallback is called every UNMUTE_INTERVAL to unmute any expired temp mutes
func unmuteCallback() {
	if db == nil {
		return
	}

	// Get all expired rows
	now := time.Now()
	rows, err := db.Query("SELECT id, discord_id, channel_id FROM mutes WHERE expiration < $1 AND expiration IS NOT NULL", now)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Println("Error querying expired tempmutes", err)
		}
		return
	}

	for rows.Next() {
		var (
			id        uuid.UUID
			discordId string
			channelId sql.NullString
		)

		err := rows.Scan(&id, &discordId, &channelId)
		if err != nil {
			log.Println("Error scanning tempmute entry", err)
			continue
		}

		// We're handling the user, so delete from db.
		// If we fail to unmute, at least we won't end up handling them forever
		rows.Close()
		_, err = db.Exec("DELETE FROM mutes WHERE id = $1", id)
		if err != nil {
			log.Println("Error deleting tempmute entry", err)
			continue
		}

		// Get channel and mute role
		var channel *discordgo.Channel
		if channelId.Valid {
			channel, err = discord.Channel(channelId.String)
			if err != nil {
				log.Println("Error getting tempmute channel from channel id "+channelId.String, err)
				continue
			}
		}
		muteRole, err := getMuteRoleForChannel(channel)
		if err != nil {
			log.Println(fmt.Sprintf("Invalid mute role for channel id \"%s\"\n", channelId.String), err)
			continue
		}

		// Construct a message to the unmuted user
		var message strings.Builder
		message.WriteString(fmt.Sprintf("Your temp mute"))
		if channel != nil {
			message.WriteString(fmt.Sprintf(" from %s", channel.Mention()))
		}
		message.WriteString(" has expired!\n")

		{ // Log the unmute
			var username = "user"
			if user, _ := discord.User(discordId); user != nil {
				username = fmt.Sprintf("@%s#%s", user.Username, user.Discriminator)
			}
			if channel == nil {
				log.Printf("Processing unmute for %s (%s) serverwide\n", username, discordId)
			} else {
				log.Printf("Processing unmute for %s (%s) from channel #%s (%s)\n", username, discordId, channel.Name, channel.ID)
			}
		}

		// Do the unmute
		err = discord.GuildMemberRoleRemove(scbc, discordId, muteRole)
		if err != nil {
			log.Println("Could not remove mute role \""+muteRole+"\" from user \""+discordId+"\"", err)
			message.WriteString("But the bot failed to unmute you! Please show this message to a moderator.\n")
		}

		// DM message to user
		dm, err := discord.UserChannelCreate(discordId)
		if err != nil {
			continue // guess we can't let em know
		}
		_ = embed(dm.ID, message.String())
	}
}

func init() {
	if db == nil {
		log.Println("WARNING: No db when initialising tempmutes callback, either rekt.go was initialised before db.go or there is no db")
	}
	go func() {
		ticker := time.NewTicker(UnmuteInterval)
		for range ticker.C {
			unmuteCallback()
		}
	}()
}

func remute(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if m.GuildID != scbc || m.User == nil {
		return
	}
	if db == nil {
		return
	}

	// Get all nonexpired rows
	now := time.Now()
	rows, err := db.Query("SELECT channel_id FROM mutes WHERE (expiration IS NULL OR expiration > $1) AND discord_id = $2", now, m.User.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Println("Error querying expired tempmutes", err)
		}
		return
	}

	for rows.Next() {
		var (
			channelId sql.NullString
		)

		err := rows.Scan(&channelId)
		if err != nil {
			log.Println("Error scanning tempmute entry", err)
			continue
		}

		// Get channel and mute role
		var channel *discordgo.Channel
		if channelId.Valid {
			channel, err = discord.Channel(channelId.String)
			if err != nil {
				log.Println("Error getting tempmute channel from channel id "+channelId.String, err)
				continue
			}
		}
		muteRole, err := getMuteRoleForChannel(channel)
		if err != nil {
			log.Println(fmt.Sprintf("Invalid mute role for channel id \"%s\"\n", channelId.String), err)
			continue
		}

		// Do the unmute
		err = discord.GuildMemberRoleAdd(scbc, m.User.ID, muteRole)
		if err != nil {
			log.Println("Could not remove mute role \""+muteRole+"\" from user \""+m.User.ID+"\"", err)
		}
	}
}
