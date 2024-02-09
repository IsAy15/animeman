package qbittorrent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
)

type (
	API struct {
		host               string
		username, password string
		client             *http.Client
	}
)

func New(ctx context.Context, host, username, password string) *API {
	client := &http.Client{
		Timeout: 3 * time.Second,
		Jar:     utils.Must(cookiejar.New(nil)),
	}
	api := &API{
		host:     fmt.Sprintf("%s/api/v2", host),
		username: username,
		password: password,
		client:   client,
	}
	api.Wait(ctx)
	if version, err := api.Version(ctx); err != nil {
		log.Fatal().Msgf("failed to connect to qBittorrent: %s", err)
	} else {
		log.Info().Msgf("connected to qBittorrent:%s", version)
	}
	return api
}

func (api *API) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	localReq := req.Clone(ctx)
	resp, err := api.client.Do(localReq)
	switch {
	case errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ECONNRESET):
		log.Warn().Msgf("qBittorrent disconnected")
		api.Wait(ctx)
		return api.Do(ctx, req)
	case err == nil && resp.StatusCode >= 400:
		if loginErr := api.Login(ctx, api.username, api.password); loginErr != nil {
			return resp, loginErr
		}
		return api.Do(ctx, req)
	}
	return resp, err
}
