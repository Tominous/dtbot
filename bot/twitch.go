package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"strings"
	"time"
)

type Twitch struct {
	Streams []*TwitchStream
	Guilds  map[string]*TwitchGuild
	DB      *DBWorker
	Conf    *Config
	Discord *discordgo.Session
}

type TwitchGuild struct {
	ID      string
	Streams []*TwitchStream
}

type TwitchStream struct {
	Login          string
	Guild          string
	Channel        string
	IsOnline       bool
	IsCustom       bool
	CustomMessage  string
	CustomImageURI string
}

type TwitchStreamResult struct {
	Data []TwitchStreamData `json:"data"`
}

type TwitchStreamData struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	GameID       string `json:"game_id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Viewers      int    `json:"viewer_count"`
	Language     string `json:"language"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type TwitchUserResult struct {
	Data []TwitchUserData `json:"data"`
}

type TwitchUserData struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	Name            string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImgURL   string `json:"profile_image_url"`
	OfflineImgURL   string `json:"offline_image_url"`
	Views           int    `json:"view_count"`
}

func TwitchInit(session *discordgo.Session, conf *Config, db *DBWorker) *Twitch {
	guilds := make(map[string]*TwitchGuild)
	var streams []*TwitchStream
	for _, g := range session.State.Guilds {
		guildStreams := db.GetTwitchStreams(g.ID)
		for _, s := range guildStreams {
			streams = append(streams, s)
		}
		guilds[g.ID] = &TwitchGuild{g.ID, guildStreams}
	}
	return &Twitch{streams, guilds, db, conf, session}
}

func (t *Twitch) Update() {
	for _, s := range t.Streams {
		timeout := time.Duration(time.Duration(1) * time.Second)
		client := &http.Client{
			Timeout: time.Duration(timeout),
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%v", s.Login), nil)
		req.Header.Add("Client-ID", t.Conf.Twitch.ClientID)
		resp, err := client.Do(req)
		var result TwitchStreamResult
		if err == nil {
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.DB.Log("Twitch", "", "Parsing Twitch API stream error")
				continue
			}
			if len(result.Data) > 0 {
				if s.IsOnline == false {
					s.IsOnline = true
					imgUrl := strings.Replace(result.Data[0].ThumbnailURL, "{width}", "720", -1)
					imgUrl = strings.Replace(imgUrl, "{height}", "480", -1)
					emb := NewEmbed(fmt.Sprintf("Hey @here %v is now live on https://www.twitch.tv/%v", result.Data[0].UserName, s.Login)).
						Field("Stream", result.Data[0].Title, true).
						AttachImgURL(imgUrl)
					_, _ = t.Discord.ChannelMessageSendEmbed(s.Channel, emb.GetEmbed())
				}
			} else {
				if s.IsOnline == true {
					s.IsOnline = false
				}
			}
			t.DB.UpdateStream(s)
		}
	}
}

func (t *Twitch) AddStream(guild, channel, login string) error {
	timeout := time.Duration(time.Duration(1) * time.Second)
	client := &http.Client{
		Timeout: time.Duration(timeout),
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", login), nil)
	req.Header.Add("Client-ID", t.Conf.Twitch.ClientID)
	resp, err := client.Do(req)
	var result TwitchUserResult
	if err == nil {
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return errors.New("parsing streamer error")
		}
		if len(result.Data) > 0 {
			stream := TwitchStream{}
			stream.Login = login
			stream.Channel = channel
			stream.Guild = guild
			t.Streams = append(t.Streams, &stream)
			t.Guilds[guild].Streams = append(t.Guilds[guild].Streams, &stream)
		}
	} else {
		return errors.New("getting streamer error")
	}
	return nil
}
