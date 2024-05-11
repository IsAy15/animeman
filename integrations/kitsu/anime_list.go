package kitsu

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error) {
	var path = API_URL + "/library-entries"
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Content-Type", "application/vnd.api+json")
	v := url.Values{
		"filter[kind]":  []string{"anime"},
		"filter[status]": []string{string(ListStatusWatching)},
		"filter[user_id]": []string{api.UserID},
		"page[limit]": []string{"20"},
		"page[offset]": []string{"0"},
	}
	req.URL.RawQuery = v.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currently watching: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch currently watching: %s", resp.Status)
	}
	var entries []AnimeListEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return convertEntry(entries), nil
}

func convertEntry(in []AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	for i := range in {
		out = append(out, animelist.Entry{
			ListStatus:  convertStatus(in[i].Status),
			// Titles:      []string{fmt.Sprint(in[i].Titles), in[i].TitleEng},
			// AiringStatus: convertAiringStatus(in[i].AiringStatus),
			// StartDate:   utils.Must(time.Parse("01-02-06", in[i].AnimeStartDateString)),
		})
	}
	return out
}


func convertStatus(in ListStatus) animelist.ListStatus {
	switch in {
	case "current":
		return animelist.ListStatusWatching
	case "planning":
		return animelist.ListStatusPlanToWatch
	case "completed":
		return animelist.ListStatusCompleted
	case "dropped":
		return animelist.ListStatusDropped
	default:
		log.Warn().Msgf("unknown status from kitsu: %s", in)
		return animelist.ListStatusAll
	}
}

func convertAiringStatus(status string) animelist.AiringStatus {
	switch status {
	case "finished":
		return animelist.AiringStatusAired
	case "current":
		return animelist.AiringStatusAiring
	default:
		log.Warn().Msgf("Unknown airing status: %s", status)
		return 0 // ou une autre valeur par défaut appropriée
	}
}
