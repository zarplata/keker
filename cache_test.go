package main

import (
	"testing"
	"time"
)

func TestCachePut(t *testing.T) {
	logger = setupLogger("ERROR")

	testToken := "testtoken"
	testMessage := "testmessage"

	cacheTTL := 4 * time.Second
	sleepTime := 2 * time.Second

	cache := newCache(cacheTTL)

	cache.put(testToken, testMessage)
	time.Sleep(sleepTime)
	objects := cache.pop(testToken)

	if testMessage != objects[0].message {
		t.Errorf(
			"incorrect message in cache, got %s, expected %s",
			objects[0].message,
			testMessage,
		)
	}
}

func TestCacheMessageEvict(t *testing.T) {
	logger = setupLogger("ERROR")

	testToken := "testtoken"
	testMessage := "testmessage"

	cacheTTL := 4 * time.Second
	sleepTime := 5 * time.Second

	cache := newCache(cacheTTL)

	cache.put(testToken, testMessage)
	time.Sleep(sleepTime)
	objects := cache.pop(testToken)

	if len(objects) != 0 {
		t.Errorf(
			"incorrect length of message cache, got %d, expected 0",
			len(objects),
		)
	}
}
