package parser

import "testing"

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
		{
			name:    "Boku no Kokoro no Yabai",
			title:   "Boku no Kokoro no Yabai Yatsu S01 1080p WEBRip DD+ x265-EMBER",
			episode: "",
			multi:   true,
		},
		{
			name:    "264 bug",
			title:   "Undead Unluck S01E13 Tatiana 1080p HULU WEB-DL AAC2.0 H 264-VARYG",
			episode: "13",
			multi:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			episode, isMulti := episodeMatch(tt.title)
			if episode != tt.episode {
				t.Errorf("episodeMatch() got episode = %v, want %v", episode, tt.episode)
			}
			if isMulti != tt.multi {
				t.Errorf("episodeMatch() got multi = %v, want %v", isMulti, tt.multi)
			}
		})
	}
}
