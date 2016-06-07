package notification

import (
	"sync"
	"time"

	"github.com/amalgam8/controller/resources"
)

// TopicName topic to send events to
const TopicName = "NewRules"

// TenantProducerCache maintains a lazy-initialized list of producers for tenant spaces.
type TenantProducerCache interface {
	StartGC()
	SendEvent(tenantID string, kafka resources.Kafka) error
	Delete(tenantID string)
}

type tenantProducerCache struct {
	client MessageHubClient
	cache  map[string]*producerCacheEntry
	ttl    time.Duration
	mutex  sync.Mutex
}

type producerCacheEntry struct {
	Producer   Producer
	Expiration time.Time
}

// NewTenantProducerCache initializes an empty cache.
func NewTenantProducerCache() TenantProducerCache {
	return &tenantProducerCache{
		cache: make(map[string]*producerCacheEntry),
		ttl:   time.Minute * 20,
	}
}

// StartGC begins a garbage collection goroutine that runs every TTL, and expires entries older than the TTL. Note that
// this means that entries can exist for longer than the TTL.
func (c *tenantProducerCache) StartGC() {
	ticker := time.NewTicker(c.ttl)
	go func() {
		select {
		case <-ticker.C:
			c.garbageCollection()
		}
	}()
}

// garbageCollection removes expired entries
func (c *tenantProducerCache) garbageCollection() {
	c.mutex.Lock() // Start modifying the cache map

	// Iterate through all cache entries, and make a list of expired entries
	toExpire := make([]string, 0, len(c.cache))
	toClose := make([]*producerCacheEntry, 0, len(c.cache))
	for tenantID, entry := range c.cache {
		if time.Now().After(entry.Expiration) {
			toExpire = append(toExpire, tenantID)
			toClose = append(toClose, entry)
		}
	}

	// Remove the expired entries from the cache
	for _, tenantID := range toExpire {
		delete(c.cache, tenantID)
	}

	c.mutex.Unlock() // Done modifying the cache map

	// Release the resources held by expired cache entries. This is can be time intensive, so we do it after unlisting
	// the entry and releasing the lock.
	for _, entry := range toClose {
		// Closes the TCP connection and releases resources. This might take a few seconds.
		entry.Producer.Close()
	}
}

// SendEvent TODO
func (c *tenantProducerCache) SendEvent(tenantID string, kafka resources.Kafka) error {
	var err error
	var expiredEntry *producerCacheEntry

	c.mutex.Lock()
	entry := c.cache[tenantID]
	if entry != nil {
		expired := time.Now().After(entry.Expiration)
		if !expired {
			err := entry.Producer.SendEvent(TopicName, tenantID, "")
			entry.Expiration = time.Now().Add(c.ttl)
			c.mutex.Unlock()
			return err
		}
		delete(c.cache, tenantID)
		expiredEntry = entry
	}
	c.mutex.Unlock()

	if expiredEntry != nil {
		expiredEntry.Producer.Close()
	}

	//FIXME
	// If there are no messagehub creds, return
	if kafka.APIKey == "" && kafka.User == "" && kafka.Password == "" &&
		kafka.AdminURL == "" && kafka.RestURL == "" &&
		len(kafka.Brokers) == 0 {
		return nil
	}

	if kafka.SASL {
		client := NewMessageHubClient(kafka.APIKey, kafka.AdminURL)

		topics, err := client.Topics()
		if err != nil {
			return err
		}

		found := false
		for _, topic := range topics {
			if topic == TopicName {
				found = true
				break
			}
		}

		if !found {
			err := client.CreateTopic(TopicName, 1) // One tenant per space, so only one partition
			if err != nil {
				return err
			}
		}
	}

	c.mutex.Lock()
	entry = c.cache[tenantID]
	if entry == nil {
		producer, err := NewProducer(ProducerConfig{
			ClientID: kafka.APIKey,
			Brokers:  kafka.Brokers,
			SASL: SASL{
				Enable:   kafka.SASL,
				User:     kafka.User,
				Password: kafka.Password,
			},
		})

		if err != nil {
			c.mutex.Unlock()
			return err
		}

		entry = &producerCacheEntry{
			Producer:   producer,
			Expiration: time.Now().Add(c.ttl),
		}
		c.cache[tenantID] = entry
	}

	err = entry.Producer.SendEvent(TopicName, tenantID, "")
	c.mutex.Unlock()

	return err
}

// Delete forces the expiration and removal of a cached producer
func (c *tenantProducerCache) Delete(tenantID string) {
	c.mutex.Lock()
	entry := c.cache[tenantID]
	delete(c.cache, tenantID)
	c.mutex.Unlock()

	if entry != nil {
		entry.Producer.Close()
	}
}
