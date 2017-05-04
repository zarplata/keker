package main

import (
	"sort"
	"time"

	"github.com/rs/xid"
)

type Cache struct {
	messageTTL time.Duration
	messages   map[string]map[string]*CacheObject
	semaphore  chan bool
}

type CacheObject struct {
	message string
	addDate int64
}

func newCache(TTL time.Duration) *Cache {
	return &Cache{
		messageTTL: TTL,
		messages:   make(map[string]map[string]*CacheObject),
		semaphore:  make(chan bool, 1),
	}
}

func (c *Cache) put(
	authToken string,
	message string,
) {

	logger.Debugf(
		"caching message %s for %s",
		message,
		authToken,
	)

	c.semaphore <- true
	messageCacheID := xid.New().String()

	messageCache := make(map[string]*CacheObject)
	if len(c.messages[authToken]) == 0 {

		messageCache[messageCacheID] = &CacheObject{
			message,
			time.Now().UnixNano(),
		}

		c.messages[authToken] = messageCache

	} else {
		messageCache = c.messages[authToken]
		messageCache[messageCacheID] = &CacheObject{
			message,
			time.Now().UnixNano(),
		}

	}

	c.messages[authToken] = messageCache

	<-c.semaphore

	go func() {
		ttlTimer := time.NewTimer(c.messageTTL)

		defer ttlTimer.Stop()

		<-ttlTimer.C

		logger.Debugf(
			"message %s for %s TTL expired, it will be evicted",
			message,
			c.messageTTL,
		)

		c.semaphore <- true
		switch len(c.messages[authToken]) {
		case 0:
			logger.Debugf(
				"messages for %s was already evicted, nothing to delete",
				authToken,
			)

			<-c.semaphore
			return
		case 1:
			delete(c.messages, authToken)
			<-c.semaphore

			logger.Debugf(
				"message cache for %s was purged",
				authToken,
			)
			return
		default:
			delete(c.messages[authToken], messageCacheID)

			<-c.semaphore
		}

		logger.Debugf(
			"message %s successfully deleted by TTL expiration",
			message,
		)
	}()

	defer func() {
		logger.Debugf(
			"message %s successfuly cached",
			message,
		)
	}()
}

func (c *Cache) pop(
	authToken string,
) []*CacheObject {

	var cachedMessages []*CacheObject

	if len(c.messages[authToken]) == 0 {
		logger.Debugf(
			"no cached messages was found for %s",
			authToken,
		)

		return cachedMessages
	}

	c.semaphore <- true

	for _, message := range c.messages[authToken] {
		cachedMessages = append(cachedMessages, message)
	}

	delete(c.messages, authToken)
	<-c.semaphore

	sort.Slice(cachedMessages, func(i, j int) bool {
		return cachedMessages[i].addDate <= cachedMessages[j].addDate
	})

	return cachedMessages
}
