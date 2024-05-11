package kitsu

type (
	ListStatus   string

	AnimeListEntry struct {
		Status ListStatus `json:"status"`
		

	}
)

const (
	ListStatusWatching  ListStatus = "current"
	ListStatusCompleted ListStatus = "completed"
	ListStatusDropped   ListStatus = "dropped"
	ListStatusPlanned  ListStatus = "planned"
	ListStatusOnHold    ListStatus = "on_hold"
)
