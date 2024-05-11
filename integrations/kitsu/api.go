package kitsu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://kitsu.io/api/edge"
const userAgent = "github.com/sonalys/animeman"

type (
    API struct {
        Username string
        UserID   string // Ajout du champ UserID
        client   *http.Client
    }
)

func New(username string) *API {
    client := &http.Client{
        Transport: roundtripper.NewUserAgentTransport(
            roundtripper.NewRateLimitedTransport(
                http.DefaultTransport, rate.NewLimiter(rate.Every(time.Second), 1),
            ), userAgent),
        Timeout: 10 * time.Second,
    }

    // Récupération du UserID
    userID, err := fetchUserID(username)
    if err != nil {
		log.Error().Msgf("failed to fetch UserID: %v", err)
        return nil
    }

    api := &API{
        client:   client,
        Username: username,
        UserID:   userID,
    }

    return api
}

// Fonction interne pour récupérer le UserID
func fetchUserID(username string) (string, error) {
    resp, err := http.Get(fmt.Sprintf("%s/users?filter[slug]=%s", API_URL, username))
    if err != nil {
        return "", fmt.Errorf("failed to fetch UserID: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to fetch UserID: %s", resp.Status)
    }

    var data struct {
        Data []struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data.Data) == 0 {
        return "", fmt.Errorf("failed to decode response or user not found")
    }

    return data.Data[0].ID, nil
}
