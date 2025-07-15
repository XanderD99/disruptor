package api

import (
	"context"
	"fmt"
	"time"

	"resty.dev/v3"

	"github.com/XanderD99/discord-disruptor/internal/errors"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

var _ store.GuildStore = (*guildStore)(nil)

type guildStore struct {
	client *resty.Client
}

type Guild struct {
	Snowflake string        `json:"snowflake"`
	Settings  GuildSettings `json:"settings"`
}

type GuildSettings struct {
	JoinChance float64 `json:"joinChance"`
	Interval   int     `json:"interval"`
	Enabled    bool    `json:"enabled"`
}

// Create implements store.GuildStore.
func (s *guildStore) Create(_ context.Context, item store.Guild) (store.Guild, error) {
	body := Response[Guild]{
		Data: Guild{
			Snowflake: item.ID,
			Settings: GuildSettings{
				JoinChance: item.Settings.Chance,
				Interval:   int(item.Settings.Interval.Seconds()),
				Enabled:    item.Settings.Enabled,
			},
		},
	}

	res := Response[Guild]{}

	resp, err := s.client.R().
		SetBody(body).
		SetResult(&res).
		Post("/guilds")

	if err != nil {
		return store.Guild{}, err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 404 {
			return store.Guild{}, errors.ErrNotFound
		}
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return store.Guild{}, errors.ErrBadRequest
		}
		return store.Guild{}, errors.ErrInternalServerError
	}

	return store.Guild{
		ID: item.ID,
		Settings: store.GuildSettings{
			Interval: time.Duration(res.Data.Settings.Interval) * time.Minute,
			Chance:   res.Data.Settings.JoinChance,
			Enabled:  res.Data.Settings.Enabled,
		},
	}, err
}

// Delete implements store.GuildStore.
func (s *guildStore) Delete(_ context.Context, id string) error {
	resp, err := s.client.R().
		SetPathParam("id", id).
		Delete("/guilds/{id}")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 404 {
			return errors.ErrNotFound
		}
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return errors.ErrBadRequest
		}
		return errors.ErrInternalServerError
	}

	return nil
}

// FindByID implements store.GuildStore.
func (s *guildStore) FindByID(ctx context.Context, id string) (store.Guild, error) {
	res := Response[Guild]{}

	resp, err := s.client.R().
		SetContext(ctx).
		SetPathParam("id", id).
		SetResult(&res).
		Get("/guilds/{id}")

	if err != nil {
		return store.Guild{}, err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 404 {
			return store.Guild{}, errors.ErrNotFound
		}
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return store.Guild{}, errors.ErrBadRequest
		}
		return store.Guild{}, errors.ErrInternalServerError
	}

	return store.Guild{
		ID: res.Data.Snowflake,
		Settings: store.GuildSettings{
			Interval: time.Minute * time.Duration(res.Data.Settings.Interval),
			Chance:   res.Data.Settings.JoinChance,
			Enabled:  res.Data.Settings.Enabled,
		},
	}, nil
}

// Update implements store.GuildStore.
func (s *guildStore) Update(ctx context.Context, id string, item store.Guild) error {
	body := Response[Guild]{
		Data: Guild{
			Snowflake: item.ID,
			Settings: GuildSettings{
				JoinChance: item.Settings.Chance,
				Interval:   int(item.Settings.Interval.Minutes()),
				Enabled:    item.Settings.Enabled,
			},
		},
	}

	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(body).
		SetPathParam("id", id).
		Put("/guilds/{id}")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 404 {
			return errors.ErrNotFound
		}
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return errors.ErrBadRequest
		}
		return errors.ErrInternalServerError
	}

	return nil
}

// Find retrieves guilds based on the provided options
// func (s *guildStore) Find(ctx context.Context, options ...store.FindOption) ([]store.Guild, int, error) {
// 	opts := store.DefaultFindOptions()
// 	opts.Apply(options...)

// 	query := opts.Filter.ToQueryString()

// 	result := &Response[[]Guild]{}
// 	request := s.client.R().
// 		SetContext(ctx).
// 		SetResult(result)

// 	if opts.Pagination != nil {
// 		request = request.
// 			SetQueryParam("pagination[limit]", fmt.Sprintf("%d", opts.Pagination.Limit)).
// 			SetQueryParam("pagination[offset]", fmt.Sprintf("%d", opts.Pagination.Offset))
// 	}

// 	if query != "" {
// 		request = request.SetQueryString(query)
// 	}

// 	resp, err := request.Get("/guilds")

// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	if resp.StatusCode() != 200 {
// 		if resp.StatusCode() == 404 {
// 			return nil, 0, errors.ErrNotFound
// 		}
// 		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
// 			return nil, 0, errors.ErrBadRequest
// 		}
// 		return nil, 0, errors.ErrInternalServerError
// 	}

// 	guilds := make([]store.Guild, len(result.Data))
// 	for i, item := range result.Data {
// 		guilds[i] = store.Guild{
// 			ID: item.Snowflake,
// 			Settings: store.GuildSettings{
// 				Interval: time.Duration(item.Settings.Interval) * time.Minute,
// 				Chance:   item.Settings.JoinChance,
// 				Enabled:  item.Settings.Enabled,
// 			},
// 		}
// 	}

// 	return guilds, result.Meta.Pagination.Total, nil
// }

// Find retrieves guilds based on the provided options
func (s *guildStore) Find(ctx context.Context, options ...store.FindOption) ([]store.Guild, int, error) {
	opts := store.DefaultFindOptions()
	opts.Apply(options...)

	// If pagination is set, use it for a single request
	if opts.Pagination != nil {
		return s.findSinglePage(ctx, opts)
	}

	// If no pagination is set, fetch all pages
	return s.findAllPages(ctx, opts)
}

// findSinglePage fetches a single page with the provided pagination
func (s *guildStore) findSinglePage(ctx context.Context, opts *store.FindOptions) ([]store.Guild, int, error) {
	query := opts.Filter.ToQueryString()

	result := &Response[[]Guild]{}
	request := s.client.R().
		SetContext(ctx).
		SetResult(result)

	request = request.
		SetQueryParam("pagination[limit]", fmt.Sprintf("%d", opts.Pagination.Limit)).
		SetQueryParam("pagination[offset]", fmt.Sprintf("%d", opts.Pagination.Offset))

	if query != "" {
		request = request.SetQueryString(query)
	}

	resp, err := request.Get("/guilds")

	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 404 {
			return nil, 0, errors.ErrNotFound
		}
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return nil, 0, errors.ErrBadRequest
		}
		return nil, 0, errors.ErrInternalServerError
	}

	guilds := make([]store.Guild, len(result.Data))
	for i, item := range result.Data {
		guilds[i] = store.Guild{
			ID: item.Snowflake,
			Settings: store.GuildSettings{
				Interval: time.Duration(item.Settings.Interval) * time.Minute,
				Chance:   item.Settings.JoinChance,
				Enabled:  item.Settings.Enabled,
			},
		}
	}

	return guilds, result.Meta.Pagination.Total, nil
}

// findAllPages fetches all pages when no pagination is specified
func (s *guildStore) findAllPages(ctx context.Context, opts *store.FindOptions) ([]store.Guild, int, error) {
	var allGuilds []store.Guild
	page := 1
	pageSize := 100
	totalCount := 0

	query := opts.Filter.ToQueryString()

	for {
		result := &Response[[]Guild]{}
		request := s.client.R().
			SetContext(ctx).
			SetResult(result).
			SetQueryParam("pagination[limit]", fmt.Sprintf("%d", pageSize)).
			SetQueryParam("pagination[offset]", fmt.Sprintf("%d", (page-1)*pageSize))

		if query != "" {
			request = request.SetQueryString(query)
		}

		resp, err := request.Get("/guilds")

		if err != nil {
			return nil, 0, err
		}

		if resp.StatusCode() != 200 {
			if resp.StatusCode() == 404 {
				return nil, 0, errors.ErrNotFound
			}
			if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
				return nil, 0, errors.ErrBadRequest
			}
			return nil, 0, errors.ErrInternalServerError
		}

		// Convert API response to store format
		guilds := make([]store.Guild, len(result.Data))
		for i, item := range result.Data {
			guilds[i] = store.Guild{
				ID: item.Snowflake,
				Settings: store.GuildSettings{
					Interval: time.Duration(item.Settings.Interval) * time.Minute,
					Chance:   item.Settings.JoinChance,
					Enabled:  item.Settings.Enabled,
				},
			}
		}

		// Add current page results to our collection
		allGuilds = append(allGuilds, guilds...)
		totalCount = result.Meta.Pagination.Total

		// Check if we've fetched all results
		if len(allGuilds) >= totalCount || len(guilds) < pageSize {
			break
		}

		page++

		// Add a small delay to avoid overwhelming the API
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Continue
		}
	}

	return allGuilds, totalCount, nil
}

// NewSoundStrapiStore creates a new Strapi store for sounds
func newGuildStore(client *resty.Client) store.GuildStore {
	client = client.Clone(context.TODO()).SetQueryParam("populate", "settings")

	return &guildStore{
		client: client,
	}
}
