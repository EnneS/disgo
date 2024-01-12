package disgo

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type (
	Player struct {
		DiscordSession *discordgo.Session

		Sessions      map[string]*Session
		sessionsMutex sync.Mutex

		EventManager *EventManager

		YoutubeClient *YoutubeClient
	}
)

func (p *Player) Init() {
	p.Sessions = make(map[string]*Session)
	p.YoutubeClient = &YoutubeClient{}
	p.EventManager = NewEventManager()

	// Add discord handler to disconnect from voice channel when bot is kicked
	p.DiscordSession.AddHandler(func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		// Delete session if bot is kicked from voice channel
		if e.UserID == s.State.User.ID && e.ChannelID == "" {
			p.deleteSession(e.GuildID, e.ChannelID)
		}

		// ===============================================

		// Delete session if bot is alone in voice channel
		guild, err := s.State.Guild(e.GuildID)
		if err != nil {
			return
		}
		var voiceChannel *discordgo.Channel
		for _, channel := range guild.VoiceStates {
			if channel.UserID == s.State.User.ID {
				voiceChannel, _ = s.State.Channel(channel.ChannelID)
				break
			}
		}
		if voiceChannel == nil {
			return
		}
		alone := true
		for _, channel := range guild.VoiceStates {
			if channel.ChannelID == voiceChannel.ID && channel.UserID != s.State.User.ID {
				alone = false
				break
			}
		}
		if alone {
			p.deleteSession(e.GuildID, e.ChannelID)
		}
	})
}

func (p *Player) Play(guildID string, chanID string, query string) (songs []Song, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s\n%s", r.(error), debug.Stack())
		}
	}()

	// Search for videos
	ytSongs, err := p.YoutubeClient.Search(query)
	if err != nil {
		return nil, err
	}

	if len(ytSongs) == 0 {
		return nil, nil
	}

	// Look for an active session
	session, err := p.getSession(guildID, chanID)
	if err != nil {
		return nil, err
	}
	queue := session.Queue

	// Add videos to queue
	for _, ytSong := range ytSongs {
		queue.AddSong(&ytSong)
	}
	session.Play()

	// Convert []Video to []Song
	songs = make([]Song, len(ytSongs))
	for i, v := range ytSongs {
		songs[i] = &v
	}

	return songs, nil
}

func (p *Player) Pause(guildID string) {
	session := p.Sessions[guildID]
	if session == nil {
		return
	}
	session.Pause()
}

func (p *Player) Resume(guildID string) {
	session := p.Sessions[guildID]
	if session == nil {
		return
	}
	session.Resume()
}

func (p *Player) Stop(guildID string) {
	session := p.Sessions[guildID]
	if session == nil {
		return
	}
	session.Stop()
}

func (p *Player) Next(guildID string) {
	session := p.Sessions[guildID]
	if session == nil {
		return
	}
	session.Next()
}

func (p *Player) On(eventName string, handler EventHandler) {
	p.EventManager.Register(eventName, handler)
}

func (p *Player) GetQueue(guildID string) *Queue {
	session := p.Sessions[guildID]
	if session == nil {
		return nil
	}
	return session.Queue
}

// ==========================
// Utils to manage sessions
// ==========================

func (p *Player) getSession(guildID string, chanID string) (*Session, error) {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()
	session := p.Sessions[guildID]
	if session == nil || !session.IsActive {
		var err error
		session, err = NewSession(guildID, chanID, p.DiscordSession, p.EventManager)
		if err != nil {
			return nil, err
		}
		p.Sessions[guildID] = session
	}
	return session, nil
}

func (p *Player) deleteSession(guildID string, chanID string) {
	p.sessionsMutex.Lock()
	defer p.sessionsMutex.Unlock()
	session := p.Sessions[guildID]
	if session == nil {
		return
	}
	session.stop()
	delete(p.Sessions, guildID)
}
