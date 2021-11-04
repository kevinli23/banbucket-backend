package appcache

import (
	"banfaucetservice/cmd/models"
	"banfaucetservice/pkg/logger"
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg/errors"
)

type Cache struct {
	Client *memcache.Client
}

func InitCache() (*Cache, error) {
	mc := memcache.New("localhost:11211")
	return &Cache{Client: mc}, nil
}

func (c *Cache) CacheDonators(value models.AllDonators) error {
	logger.Info.Println("Caching donators")
	enc, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Client.Set(&memcache.Item{Key: "donators", Value: enc, Expiration: 60 * 60 * 24})
}

func (c *Cache) GetDonators() ([]models.Donator, error) {
	fetchItem, err := c.Client.Get("donators")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch donators from memcache")
	}

	var donators models.AllDonators

	err = json.Unmarshal(fetchItem.Value, &donators)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal donators from memcache")
	}

	return donators.Donators, nil
}
