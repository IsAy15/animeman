package parser

import (
	"testing"
)

func Test_matchEpisode(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		episode string
		multi   bool
	}{
		{name: "0x15", title: "Frieren 0x15", episode: "15", multi: false},
		{name: "-15", title: "Frieren - 15", episode: "15", multi: false},
		{name: "S02E15", title: "Frieren S02E15", episode: "15", multi: false},
		{name: "Season", title: "Frieren Season 2", episode: "", multi: true},
		{name: "Season with episode", title: "Frieren Season 2 - 15", episode: "15", multi: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			episode, isMulti := matchEpisode(tt.title)
			if episode != tt.episode {
				t.Errorf("matchEpisode() got episode = %v, want %v", episode, tt.episode)
			}
			if isMulti != tt.multi {
				t.Errorf("matchEpisode() got multi = %v, want %v", isMulti, tt.multi)
			}
		})
	}
}

func Test_matchSeason(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "Tatami Galaxy",
			title: "[EMBER] The Tatami Galaxy (2010) (Season 1) [BDRip] [1080p HEVC 10 bits] (Yojouhan Shinwa Taikei)",
			want:  "1",
		},
		{name: "2x15", title: "Frieren 2x15", want: "2"},
		{name: "S02E15", title: "Frieren S02E15", want: "2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchSeason(tt.title); got != tt.want {
				t.Errorf("matchSeason() = season '%v', want '%v'", got, tt.want)
			}
		})
	}
}