package main

import (
	"math/rand"
	"time"
)

type (
	Song interface {
		URL() (string, error)
		Title() string
		Author() string
		Duration() string
	}

	Queue struct {
		songs []Song
	}
)

func (q *Queue) AddSong(s Song) {
	q.songs = append(q.songs, s)
}

func (q *Queue) RemoveSong(i int) {
	q.songs = append(q.songs[:i], q.songs[i+1:]...)
}

func (q *Queue) GetSongs() []Song {
	return q.songs
}

func (q *Queue) GetSong(i int) Song {
	return q.songs[i]
}

func (q *Queue) Len() int {
	return len(q.songs)
}

func (q *Queue) IsEmpty() bool {
	return q.Len() == 0
}

func (q *Queue) Clear() {
	q.songs = []Song{}
}

func (q *Queue) Pop() Song {
	song := q.songs[0]
	q.songs = q.songs[1:]
	return song
}

func (q *Queue) MoveSong(i int, j int) {
	song := q.songs[i]
	q.songs = append(q.songs[:i], q.songs[i+1:]...)
	q.songs = append(q.songs[:j], append([]Song{song}, q.songs[j:]...)...)
}

func (q *Queue) Shuffle() {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	r.Shuffle(len(q.songs), func(i, j int) {
		q.songs[i], q.songs[j] = q.songs[j], q.songs[i]
	})
}
