package cmd

import (
	"strings"
	"time"

	"github.com/FlameInTheDark/dtbot/bot"
)

// DebugCommand special bot commands handler
func DebugCommand(ctx bot.Context) {
	ctx.MetricsCommand("debug", "admin")
	if ctx.GetRoles().ExistsName("bot.admin") {
		if len(ctx.Args) == 0 {
			return
		}
		switch ctx.Args[0] {
		case "roles":
			var roles []string
			for _, val := range ctx.GetRoles().Roles {
				roles = append(roles, val.Name)
			}
			ctx.ReplyEmbedPM("Debug", strings.Join(roles, ", "))
		case "time":
			ctx.ReplyEmbedPM("Debug", time.Now().String())
		case "session":
			sess := ctx.Sessions.GetByGuild(ctx.Guild.ID)
			if sess != nil {
				ctx.ReplyEmbed("Debug", sess.ChannelID)
			} else {
				ctx.ReplyEmbed("Debug", "Session is nil")
			}
		}
	} else {
		ctx.ReplyEmbedPM("Debug", "Not a Admin")
	}
}
