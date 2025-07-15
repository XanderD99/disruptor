package api

import (
	"context"
	"fmt"

	"resty.dev/v3"

	"github.com/XanderD99/discord-disruptor/internal/errors"
	"github.com/XanderD99/discord-disruptor/internal/store"
)

type soundStore struct {
	client *resty.Client
}

// newSoundStore creates a new Strapi store for sounds
func newSoundStore(client *resty.Client) store.SoundStore {
	client = client.Clone(context.TODO()).SetQueryParam("populate", "file")

	return &soundStore{
		client: client,
	}
}

type Sound struct {
	ID   string `json:"documentId"`
	Name string `json:"name"`
	File File   `json:"file"`
}

type File struct {
	URL string `json:"url"`
}

// Find retrieves sounds based on the provided options
func (s *soundStore) Find(ctx context.Context, opts ...store.FindOption) ([]store.Sound, error) {
	options := store.DefaultFindOptions()
	options.Apply(opts...)

	query := make(map[string]string)
	// if options.Limit > 0 {
	//   query["pagination[limit]"] = fmt.Sprint(options.Limit)
	// }
	// if options.Offset > 0 {
	//   query["pagination[start]"] = fmt.Sprint(options.Offset)
	// }
	// if options.Sort != nil {
	//   query["sort"] = fmt.Sprint(options.Sort) // Assuming Sort is a string or can be converted to a string
	// }

	// if filter, ok := options.Filter.(*StrapiFilter); ok {
	//   // Assuming StrapiFilter is a map[string]string
	//   maps.Copy(query, filter.ToQueryParams())
	// }
	res := &Response[[]Sound]{}
	resp, err := s.client.R().
		SetContext(ctx).
		SetQueryParams(query).
		SetResult(&res).
		Get("/sounds")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return nil, errors.ErrBadRequest
		}
		return nil, errors.ErrInternalServerError
	}

	sounds := make([]store.Sound, len(res.Data))
	for i, sound := range res.Data {
		sounds[i] = store.Sound{
			ID:   sound.ID, // Assuming Name is unique and can be used as ID
			Name: sound.Name,
			URL:  sound.File.URL,
		}
	}
	return sounds, nil
}

// FindByID retrieves a sound by its ID
func (s *soundStore) FindByID(ctx context.Context, id string) (store.Sound, error) {
	res := &Response[Sound]{}
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&res).
		SetPathParam("id", id).
		Get("/sounds/{id}")

	if err != nil {
		return store.Sound{}, err
	}

	if resp.StatusCode() != 200 {
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return store.Sound{}, fmt.Errorf("sound with id %s not found: %w", id, errors.ErrNotFound)
		}
		return store.Sound{}, fmt.Errorf("failed to retrieve sound with id %s: %w", id, errors.ErrInternalServerError)
	}

	return store.Sound{
		ID:   res.Data.ID,
		Name: res.Data.Name,
		URL:  res.Data.File.URL,
	}, nil
}

// Random retrieves a random sound
func (s *soundStore) Random(ctx context.Context) (store.Sound, error) {
	return s.FindByID(ctx, "random")
}
