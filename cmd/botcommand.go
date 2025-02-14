package cmd

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo/bson"
	"strconv"
	"strings"

	"github.com/FlameInTheDark/dtbot/bot"
)

func showLogs(ctx *bot.Context, count int) {
	logs := ctx.DB.LogGet(count)
	var logString = []string{""}
	for _, log := range logs {
		logString = append(logString, fmt.Sprintf("[%v] %v: %v\n", log.Date, log.Module, log.Text))
	}
	ctx.ReplyEmbedPM("Logs", strings.Join(logString, ""))
}

// BotCommand special bot commands handler
func BotCommand(ctx bot.Context) {
	if ctx.IsServerAdmin() {
		if len(ctx.Args) == 0 {
			return
		}
		switch ctx.Args[0] {
		case "clear":
			ctx.MetricsCommand("bot", "clear")
			if len(ctx.Args) < 2 {
				ctx.BotMsg.Clear(&ctx, 0)
				return
			}
			from, err := strconv.Atoi(ctx.Args[1])
			if err != nil {
				return
			}
			ctx.BotMsg.Clear(&ctx, from)
		case "logs":
			ctx.MetricsCommand("bot", "logs")
			if ctx.IsAdmin() {
				if len(ctx.Args) < 2 {
					showLogs(&ctx, 10)
				} else {
					count, err := strconv.Atoi(ctx.Args[1])
					if err != nil {
						fmt.Println(err)
						return
					}
					showLogs(&ctx, count)
				}
			}
		case "conflist":
			ctx.MetricsCommand("bot", "conflist")
			ctx.ReplyEmbed("Config", ctx.Loc("conf_list"))
		case "setconf":
			botSetConf(&ctx)
		case "stations":
			botStations(&ctx)
		case "guild":
			botGuild(&ctx)
		case "stats":
			ctx.MetricsCommand("bot", "stats")
			if !ctx.IsAdmin() {
				return
			}
			var users int
			for _, g := range ctx.Discord.State.Guilds {
				users += len(g.Members)
			}
			ctx.ReplyEmbed("Stats", fmt.Sprintf(ctx.Loc("stats_command"), len(ctx.Discord.State.Guilds), users))
		case "blacklist":
			botBlacklist(&ctx)
		}
	} else {
		ctx.ReplyEmbed("Bot", ctx.Loc("admin_require"))
	}
}

func guildsListID(guilds []*discordgo.Guild, current, pages int) string {
	var list string
	for _, g := range guilds {
		var gName string
		if len(g.Name) > 20 {
			gName = fmt.Sprintf("%v...", g.Name[:20])
		} else {
			gName = g.Name
		}
		list += fmt.Sprintf("[%v] - %v\n", g.ID, gName)
	}
	list += fmt.Sprintf("Page: %v | %v", current, pages)
	return list
}

func guildsListName(guilds []*discordgo.Guild, current, pages int) string {
	var list string
	for i, g := range guilds {
		var gName string
		if len(g.Name) > 20 {
			gName = fmt.Sprintf("%v...", g.Name[:20])
		} else {
			gName = g.Name
		}
		list += fmt.Sprintf("[%v] - %v | U: %v\n", i, gName, len(g.Members))
	}
	list += fmt.Sprintf("Page: %v | %v", current, pages)
	return list
}

func botSetConf(ctx *bot.Context) {
	ctx.MetricsCommand("bot", "setconf")
	if len(ctx.Args) > 2 {
		target := strings.Split(ctx.Args[1], ".")
		switch target[0] {
		case "general":
			switch target[1] {
			case "language":
				ctx.Guilds.Guilds[ctx.Guild.ID].Language = ctx.Args[2]
				_ = ctx.DB.Guilds().Update(bson.M{"id": ctx.Guild.ID}, bson.M{"$set": bson.M{"language": ctx.Args[2]}})
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("Language set to: %v", ctx.Args[2]))
			case "timezone":
				tz, err := strconv.Atoi(ctx.Args[1])
				if err != nil {
					ctx.ReplyEmbedPM("Settings", "Not a number")
				}
				ctx.Guilds.Guilds[ctx.Guild.ID].Timezone = tz
				_ = ctx.DB.Guilds().Update(bson.M{"id": ctx.Guild.ID}, bson.M{"$set": bson.M{"timezone": tz}})
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("Timezone set to: %v", ctx.Args[2]))
			case "nick":
				_ = ctx.Discord.GuildMemberNickname(ctx.Guild.ID, "@me", ctx.Args[2])
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("Nickname changed to %v", ctx.Args[2]))
			}
		case "weather":
			switch target[1] {
			case "city":
				ctx.Guilds.Guilds[ctx.Guild.ID].WeatherCity = ctx.Args[2]
				_ = ctx.DB.Guilds().Update(bson.M{"id": ctx.Guild.ID}, bson.M{"$set": bson.M{"weathercity": ctx.Args[2]}})
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("Weather city set to: %v", ctx.Args[2]))
			}
		case "news":
			switch target[1] {
			case "country":
				ctx.Guilds.Guilds[ctx.Guild.ID].NewsCounty = ctx.Args[2]
				_ = ctx.DB.Guilds().Update(bson.M{"id": ctx.Guild.ID}, bson.M{"$set": bson.M{"weathercountry": ctx.Args[2]}})
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("News country set to: %v", ctx.Args[2]))
			}
		case "embed":
			switch target[1] {
			case "color":
				var color int64
				var err error
				if strings.HasPrefix(ctx.Args[2], "#") {
					color, err = strconv.ParseInt(ctx.Args[2][1:], 16, 32)
					if err != nil {
						ctx.Log("Config", ctx.Guild.ID, fmt.Sprintf("error setting parameter %v to value %v: %v", ctx.Args[1], target[2], err.Error()))
						return
					}
				} else {
					color, err = strconv.ParseInt(ctx.Args[2], 16, 32)
					if err != nil {
						ctx.Log("Config", ctx.Guild.ID, fmt.Sprintf("error setting parameter %v to value %v: %v", ctx.Args[1], ctx.Args[2], err.Error()))
						return
					}
				}
				ctx.Guilds.Guilds[ctx.Guild.ID].EmbedColor = int(color)
				_ = ctx.DB.Guilds().Update(bson.M{"id": ctx.Guild.ID}, bson.M{"$set": bson.M{"embedcolor": int(color)}})
				ctx.ReplyEmbedPM("Config", fmt.Sprintf("Embed color set to: %v", ctx.Args[2]))
			}
		}
	}
}

func botGuild(ctx *bot.Context) {
	ctx.MetricsCommand("bot", "guild")
	if len(ctx.Args) < 2 && !ctx.IsAdmin() {
		return
	}
	switch ctx.Arg(1) {
	case "leave":
		if len(ctx.Args) < 3 {
			return
		}
		err := ctx.Discord.GuildLeave(ctx.Args[2])
		if err != nil {
			ctx.Log("Guild", ctx.Guild.ID, fmt.Sprintf("error leaving from guild [%v]: %v", ctx.Args[2], err.Error()))
			ctx.ReplyEmbedPM("Guild", fmt.Sprintf("Error leaving from guild [%v]: %v", ctx.Args[2], err.Error()))
			return
		}
		ctx.ReplyEmbedPM("Guild", fmt.Sprintf("Leave from guild: %v", ctx.Args[2]))
	case "info":
		var guild *discordgo.Guild
		var err error

		if len(ctx.Args) < 3 {
			guild = ctx.Guild
		} else {
			guild, err = ctx.Discord.Guild(ctx.Args[2])
			if err != nil {
				ctx.ReplyEmbed(ctx.Loc("guild_info"), ctx.Loc("guild_not_found"))
				return
			}
		}

		var (
			usersOnline   int
			usersOffline  int
			usersIdle     int
			usersDND      int
			usersBot      int
			channelsVoice int
			channelsText  int
			guildEmojis   int
			guildRoles    string
			guildUsers    int
			guildOwner    string
		)

		for _, m := range guild.Members {
			if m.User.ID == guild.OwnerID {
				guildOwner = m.User.Username + "#" + m.User.Discriminator
			}
		}

		for _, p := range guild.Presences {
			switch p.Status {
			case discordgo.StatusOnline:
				usersOnline++
			case discordgo.StatusIdle:
				usersIdle++
			case discordgo.StatusDoNotDisturb:
				usersDND++
			}
		}
		guildUsers = len(guild.Members)
		usersOffline = guild.MemberCount - (usersOnline + usersIdle + usersDND)

		for _, m := range guild.Members {
			if m.User.Bot {
				usersBot++
			}
		}

		for _, c := range guild.Channels {
			switch c.Type {
			case discordgo.ChannelTypeGuildText:
				channelsText++
			case discordgo.ChannelTypeGuildVoice:
				channelsVoice++
			}
		}

		for _, _ = range guild.Emojis {
			guildEmojis++
		}

		for _, r := range guild.Roles {
			if r.Name == "@everyone" {
				continue
			}
			if len(guildRoles+r.Name) < 100 {
				guildRoles += r.Name + "\n"
			}
		}

		emb := bot.NewEmbed(ctx.Loc("guild_info"))
		emb.Color(ctx.GetGuild().EmbedColor)
		emb.Field(ctx.Loc("guild_name"), guild.Name, true)
		emb.Field(ctx.Loc("guild_emoji"), fmt.Sprintf(ctx.Loc("guild_emoji_count"), guildEmojis), true)
		emb.Field(ctx.Loc("guild_channels"), fmt.Sprintf(ctx.Loc("guild_channels_format"), channelsText, channelsVoice), true)
		emb.Field(ctx.Loc("guild_id"), guild.ID, true)
		emb.Field(ctx.Loc("guild_users"), fmt.Sprintf(ctx.Loc("guild_users_format"), guildUsers, usersOnline, usersOffline, usersIdle, usersDND, usersBot), true)
		emb.Field(ctx.Loc("guild_roles"), guildRoles, true)
		emb.Field(ctx.Loc("guild_owner"), fmt.Sprintf(ctx.Loc("guild_owner_format"), guildOwner, guild.OwnerID), true)
		emb.Send(ctx)
	case "list":
		var selected string
		var paged = false
		if len(ctx.Args) > 2 && ctx.Args[2] == "id" {
			if len(ctx.Args) > 3 {
				selected = ctx.Args[3]
				paged = true
			} else {
				selected = "1"
			}
		} else {
			if len(ctx.Args) > 2 {
				selected = ctx.Args[2]
				paged = true
			} else {
				selected = "1"
			}
		}
		// calculates count of pages
		guilds := ctx.Discord.State.Guilds
		pages := 1 + int(len(guilds)/20)
		// paginate
		var indexTo = 20
		if paged {
			page, err := strconv.Atoi(selected)
			if err == nil {
				indexTo = page * 20
				indexFrom := indexTo - 20

				if indexFrom < 0 {
					indexFrom = 0
				}
				if indexTo > len(guilds) {
					indexTo = len(guilds) - 1
				}
				if len(ctx.Args) > 2 && ctx.Args[2] == "id" {
					ctx.ReplyEmbed("Guilds", guildsListID(guilds[indexFrom:indexTo], page, pages)+fmt.Sprintf("\nFrom: %v\nTo: %v", indexFrom, indexTo))
				} else {
					ctx.ReplyEmbed("Guilds", guildsListName(guilds[indexFrom:indexTo], page, pages)+fmt.Sprintf("\nFrom: %v\nTo: %v", indexFrom, indexTo))
				}

			} else {
				ctx.ReplyEmbed("Guilds", fmt.Sprintf("Selected: %v\nError: %v", selected, err.Error()))
			}

		} else {
			if indexTo > len(guilds) {
				indexTo = len(guilds) - 1
			}
			if len(ctx.Args) > 2 && ctx.Args[2] == "id" {
				ctx.ReplyEmbed("Guilds", guildsListID(guilds[:indexTo], 1, 1)+fmt.Sprintf("\nTo: %v", indexTo))
			} else {
				ctx.ReplyEmbed("Guilds", guildsListName(guilds[:indexTo], 1, 1)+fmt.Sprintf("\nTo: %v", indexTo))
			}
		}
	}
}

func botStations(ctx *bot.Context) {
	ctx.MetricsCommand("bot", "stations")
	if !ctx.IsAdmin() {
		return
	}
	switch ctx.Args[1] {
	case "add":
		if len(ctx.Args) > 5 {
			name := strings.Join(ctx.Args[5:], " ")
			err := ctx.DB.AddRadioStation(name, ctx.Args[3], ctx.Args[4], ctx.Args[2])
			if err != nil {
				ctx.ReplyEmbed("Stations", "Adding error")
			}
			ctx.ReplyEmbed("Stations", ctx.Loc("stations_added"))
		} else {
			ctx.ReplyEmbed("Stations", "Arguments missed")
		}
	case "remove":
		if len(ctx.Args) > 2 {
			err := ctx.DB.RemoveRadioStation(ctx.Args[2])
			if err != nil {
				ctx.ReplyEmbed("Stations", ctx.Loc("stations_removed"))
			}
		} else {
			ctx.ReplyEmbed("Stations", "Arguments missed")
		}
	}
}

func botBlacklist(ctx *bot.Context) {
	if len(ctx.Args) > 2 {
		switch ctx.Args[1] {
		case "addguild":
			ctx.BlacklistAddGuild(ctx.Args[2])
			ctx.ReplyEmbed("Bot", fmt.Sprintf(ctx.Loc("blacklist_guild_add"), ctx.Args[2]))
		case "adduser":
			ctx.BlacklistAddUser(ctx.Args[2])
			ctx.ReplyEmbed("Bot", fmt.Sprintf(ctx.Loc("blacklist_user_add"), ctx.Args[2]))
		case "removeuser":
			ctx.BlacklistRemoveUser(ctx.Args[2])
			ctx.ReplyEmbed("Bot", fmt.Sprintf(ctx.Loc("blacklist_user_remove"), ctx.Args[2]))
		case "removeguild":
			ctx.BlacklistRemoveGuild(ctx.Args[2])
			ctx.ReplyEmbed("Bot", fmt.Sprintf(ctx.Loc("blacklist_guild_remove"), ctx.Args[2]))
		}
	}

}