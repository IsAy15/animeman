package discovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/parser"
)

type (
	Config struct {
		Sources          []string
		Qualitites       []string
		Category         string
		DownloadPath     string
		CreateShowFolder bool
		PollFrequency    time.Duration
	}

	Dependencies struct {
		MAL    *myanimelist.API
		NYAA   *nyaa.API
		QB     *qbittorrent.API
		Config Config
	}

	Controller struct {
		dep Dependencies
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep: dep,
	}
}

func (c *Controller) Start(ctx context.Context) error {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())
	timer := time.NewTicker(c.dep.Config.PollFrequency)
	defer timer.Stop()
	c.RunDiscovery(ctx)
	for {
		select {
		case <-timer.C:
			err := c.RunDiscovery(ctx)
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Error().Msgf("scan failed: %s", err)
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Controller) RunDiscovery(ctx context.Context) error {
	log.Info().Msg("discovery started")
	entries, err := c.dep.MAL.GetAnimeList(ctx,
		myanimelist.ListStatusWatching,
	)
	if err != nil {
		panic(fmt.Errorf("getting MAL list: %w", err))
	}
	log.Info().Msgf("processing %d entries from MAL", len(entries))
	var addedCount int
	for _, entry := range entries {
		log.Debug().Msgf("Digesting entry '%s'", entry.GetTitle())
		torrents, err := c.dep.NYAA.List(ctx,
			nyaa.CategoryAnimeEnglishTranslated,
			nyaa.Query(parser.StripTitle(entry.GetTitle())),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Sources, "|"))),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Qualitites, "|"))),
		)
		log.Debug().Str("entry", entry.GetTitle()).Msgf("Found %d torrents", len(torrents))
		if err != nil {
			return fmt.Errorf("getting nyaa list: %w", err)
		}
		added, err := c.digestEntry(ctx, entry, torrents)
		if err != nil {
			if errors.Is(err, qbittorrent.ErrUnauthorized) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("failed to digest entry: %w", err)
			}
			continue
		}
		if added {
			addedCount++
		}
	}
	if addedCount > 0 {
		log.Info().Msgf("added %d torrents", addedCount)
	}
	return nil
}

func (c *Controller) digestEntry(ctx context.Context, entry myanimelist.AnimeListEntry, torrents []nyaa.Entry) (bool, error) {
	if len(torrents) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.GetTitle())
		return false, nil
	}
	var torrent nyaa.Entry
	var tags qbittorrent.Tags
	for i := range torrents {
		torrent = torrents[i]
		log.Debug().Str("entry", entry.GetTitle()).Msgf("Analyzing torrent '%s'", torrent.Title)
		parsedTitle := parser.ParseTitle(torrent.Title)
		if parsedTitle.IsMultiEpisode && entry.AiringStatus == myanimelist.AiringStatusAiring {
			log.Debug().Str("entry", entry.GetTitle()).Msgf("torrent '%s' dropped: multi-episode for currently airing", torrent.Title)
			continue
		}
		tags = qbittorrent.Tags{"animeman", entry.GetTitle(), fmt.Sprintf("S%sE%s", parsedTitle.Season, parsedTitle.Episode)}
		// check if torrent already exists, if so we skip it.
		torrentList, err := c.dep.QB.List(ctx, tags)
		if err != nil {
			return false, fmt.Errorf("listing torrents: %w", err)
		}
		if len(torrentList) > 0 {
			log.Debug().Str("entry", entry.GetTitle()).Msgf("S%sE%s already exists for %s in qBitTorrent client", parsedTitle.Season, parsedTitle.Episode, entry.GetTitle())
			return false, nil
		}
	}
	var savePath qbittorrent.SavePath
	if c.dep.Config.CreateShowFolder {
		savePath = qbittorrent.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
	} else {
		savePath = qbittorrent.SavePath(c.dep.Config.DownloadPath)
	}
	err := c.dep.QB.AddTorrent(ctx,
		tags,
		savePath,
		qbittorrent.TorrentURL{torrent.Link},
		qbittorrent.Category(c.dep.Config.Category),
	)
	if err != nil {
		return false, fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().
		Str("savePath", string(savePath)).
		Msgf("torrent '%s' added", entry.GetTitle())
	return true, nil
}
