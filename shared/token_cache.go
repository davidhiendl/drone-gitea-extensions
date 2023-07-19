package shared

import (
	"code.gitea.io/sdk/gitea"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
	"sync"
	"time"
)

func NewTokenCache(client *gitea.Client, config *AppConfig) TokenCache {

	giteaUrl, err := url.Parse(config.GiteaURL)
	if err != nil {
		logrus.Fatalf("failed to parse gitea url: %+v", err)
	}

	return TokenCache{
		entries:                 map[int64]*tokenCacheEntry{},
		client:                  client,
		config:                  config,
		giteaPackagesURL:        config.GiteaURL + "/api/packages",
		giteaURL:                config.GiteaURL,
		giteaDockerRegistryHost: giteaUrl.Hostname(),
	}
}

type TokenCache struct {
	mu                      sync.Mutex
	entries                 map[int64]*tokenCacheEntry
	client                  *gitea.Client
	config                  *AppConfig
	giteaPackagesURL        string
	giteaURL                string
	giteaDockerRegistryHost string
	gcJob                   bool
}

type tokenCacheEntry struct {
	createdAt int64
	buildId   int64
	token     *gitea.AccessToken
}

func (c *TokenCache) StartCleanupAccessTokenJob() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gcJob {
		return
	}
	c.gcJob = true

	go func() {
		c.CleanupAccessTokens()
		time.Sleep(65 * time.Second)
	}()
}

func (c *TokenCache) StopCleanupAccessTokenJob() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gcJob = false
}

func (c *TokenCache) CleanupAccessTokens() {
	c.mu.Lock()
	defer c.mu.Unlock()

	ts := time.Now().Unix()
	for key, value := range c.entries {
		if value.createdAt+int64(c.config.GiteaDroneTokenTTL) > ts {
			delete(c.entries, key)
		}
	}
}

func (c *TokenCache) GetAccessToken(buildId int64, sender string) (*gitea.AccessToken, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// attempt to fetch token from cache
	cacheEntry, exists := c.entries[buildId]
	if exists {
		logrus.Debugf("reusing cached token for buildId=%d sender=%s", buildId, sender)
		return cacheEntry.token, nil
	}

	if len(sender) == 0 {
		return nil, errors.New(fmt.Sprintf("build is missing sender info: buildId=%d", buildId))
	}

	c.client.SetSudo("")
	_, _, err := c.client.GetUserInfo(sender)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to find gitea user: '%s'", sender))
	}

	now := time.Now().Unix()
	accessTokenName := fmt.Sprintf("%s_%d_%d", c.config.GiteaDroneTokenPrefix, buildId, now)

	c.client.SetSudo(sender)
	token, _, err := c.client.CreateAccessToken(gitea.CreateAccessTokenOption{
		Name: accessTokenName,
		Scopes: []gitea.AccessTokenScope{
			gitea.AccessTokenScopeAll,
		},
	})
	if err != nil {
		// TODO handle already exists errors?
		return nil, errors.New(fmt.Sprintf("failed to create gitea access token: err=%+v buildId=%s sender=%s", err, buildId, sender))
	}

	// store token for re-use
	c.entries[buildId] = &tokenCacheEntry{
		buildId:   buildId,
		createdAt: now,
		token:     token,
	}

	return token, nil
}
