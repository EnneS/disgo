package main

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/ennes/dca"
)

type (
	AudioStream struct {
		Voice            *discordgo.VoiceConnection
		streamingSession *dca.StreamingSession
		done             chan error
	}
)

const (
	CHANNELS   int = 2
	FRAME_RATE int = 48000
	FRAME_SIZE int = 960
	MAX_BYTES  int = (FRAME_SIZE * 2) * 2
)

func NewAudioStream(vc *discordgo.VoiceConnection) *AudioStream {
	return &AudioStream{
		Voice: vc,
		done:  make(chan error),
	}
}

func (audioS *AudioStream) Play(s Song) error {
	url, err := s.URL()
	if err != nil {
		return err
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"
	encodingSession, err := dca.EncodeFile(url, options)
	if err != nil {
		return err
	}
	defer encodingSession.Cleanup()

	audioS.streamingSession = dca.NewStream(encodingSession, audioS.Voice, audioS.done)
	err = <-audioS.done
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (audioS *AudioStream) IsRunning() bool {
	if audioS.streamingSession == nil {
		return false
	}
	finished, _ := audioS.streamingSession.Finished()
	return !finished
}

func (audioS *AudioStream) Pause() {
	audioS.streamingSession.SetPaused(true)
}

func (audioS *AudioStream) Resume() {
	audioS.streamingSession.SetPaused(false)
}

func (audioS *AudioStream) Stop() {
	if !audioS.IsRunning() {
		return
	}
	audioS.done <- io.EOF         // By sending EOF, the stream will stop
	audioS.streamingSession = nil // Clean up the session
}
