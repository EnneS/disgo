package disgo

import (
	"github.com/bwmarrin/discordgo"
)

type (
	Session struct {
		Queue          *Queue
		IsActive       bool
		guildId        string
		channelId      string
		eventManager   *EventManager
		vc             *discordgo.VoiceConnection
		audioStream    *AudioStream
		queueScheduler chan schedulerEvent
	}

	schedulerEvent int
)

const (
	SCHEDULER_EVENT_PLAY schedulerEvent = iota
	SCHEDULER_EVENT_PAUSE
	SCHEDULER_EVENT_RESUME
	SCHEDULER_EVENT_STOP
	SCHEDULER_EVENT_NEXT
	SCHEDULER_EVENT_SHUFFLE
	SCHEDULER_EVENT_ERROR
)

func NewSession(guildId string, channelId string, d *discordgo.Session, em *EventManager) (*Session, error) {
	vc, err := d.ChannelVoiceJoin(guildId, channelId, false, true)
	vc.LogLevel = discordgo.LogWarning
	if err != nil {
		return nil, err
	}
	session := &Session{
		Queue:          &Queue{},
		IsActive:       true,
		guildId:        guildId,
		channelId:      channelId,
		vc:             vc,
		audioStream:    NewAudioStream(vc),
		eventManager:   em,
		queueScheduler: make(chan schedulerEvent),
	}
	go session.scheduler()
	return session, err
}

func (s *Session) Play() {
	s.queueScheduler <- SCHEDULER_EVENT_PLAY
}

func (s *Session) Pause() {
	s.queueScheduler <- SCHEDULER_EVENT_PAUSE
}

func (s *Session) Resume() {
	s.queueScheduler <- SCHEDULER_EVENT_RESUME
}

func (s *Session) Stop() {
	s.queueScheduler <- SCHEDULER_EVENT_STOP
}

func (s *Session) Next() {
	s.queueScheduler <- SCHEDULER_EVENT_NEXT
}

func (s *Session) Shuffle() {
	s.queueScheduler <- SCHEDULER_EVENT_SHUFFLE
}

func (s *Session) scheduler() {
	for {
		event := <-s.queueScheduler
		switch event {
		case SCHEDULER_EVENT_PLAY:
			go s.playNext()
		case SCHEDULER_EVENT_PAUSE:
			s.audioStream.Pause()
			s.eventManager.Push(&PauseEvent{})
		case SCHEDULER_EVENT_RESUME:
			s.audioStream.Resume()
			s.eventManager.Push(&ResumeEvent{})
		case SCHEDULER_EVENT_STOP:
			s.stop()
		case SCHEDULER_EVENT_NEXT:
			s.audioStream.Stop() // Stop the stream which will trigger the next song to play
		case SCHEDULER_EVENT_SHUFFLE:
			s.Queue.Shuffle()
		case SCHEDULER_EVENT_ERROR:
			s.stop() // Cleanup the session and disconnect
		}
	}
}

func (s *Session) playNext() {
	if s.Queue.IsEmpty() { // No more songs to play
		s.stop()
		return
	}

	// Check if already playing
	if s.audioStream.IsRunning() {
		return
	}

	song := s.Queue.Pop()
	s.eventManager.Push(&PlayEvent{
		Song: song,
	})
	err := s.audioStream.Play(song)
	if err != nil {
		s.queueScheduler <- SCHEDULER_EVENT_ERROR
		return
	}
	s.Play() // Play next song
}

func (s *Session) stop() {
	if !s.IsActive {
		return
	}
	s.audioStream.Stop()
	s.Queue.Clear()
	s.IsActive = false
	s.vc.Disconnect()
	s.eventManager.Push(&StopEvent{})
}
