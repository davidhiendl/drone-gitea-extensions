package shared

import (
	"code.gitea.io/sdk/gitea"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func StartGiteaTokenCleanupBackgroundJob(client *gitea.Client, cfg *AppConfig) {
	go GiteaTokenCleanupBackgroundJob(client, cfg)
}

func GiteaTokenCleanupBackgroundJob(client *gitea.Client, cfg *AppConfig) {
	// between 0-60s on startup
	time.Sleep(time.Duration(rand.Int31n(60)) * time.Second)
	for {
		logrus.Infof("Running gitea token cleanup ...")
		giteaTokenCleanup(client, cfg)
		logrus.Infof("Finished gitea token cleanup.")
		time.Sleep(5*time.Minute + time.Duration(rand.Int31n(60))*time.Second)
	}
}

func giteaTokenCleanup(client *gitea.Client, cfg *AppConfig) {

	pageSize := 100

	usersPage := 0
	for {
		client.SetSudo("")
		users, _, err := client.AdminListUsers(gitea.AdminListUsersOptions{ListOptions: gitea.ListOptions{
			Page:     usersPage,
			PageSize: pageSize,
		}})

		if err != nil {
			logrus.Errorf("Failed to list users got token cleanup")
			return
		}

		for _, user := range users {
			client.SetSudo(user.UserName)

			tokensPage := 0
			for {
				tokens, _, err := client.ListAccessTokens(gitea.ListAccessTokensOptions{ListOptions: gitea.ListOptions{
					Page:     tokensPage,
					PageSize: pageSize,
				}})

				if err != nil {
					logrus.Errorf("Failed to list user '%s' tokens for cleanup", user.UserName)
					break
				}

				for _, token := range tokens {
					giteaCheckAndCleanUserToken(client, cfg, user, token)
				}

				// incomplete page indicates there are no more users
				if len(tokens) < pageSize {
					break
				}
				tokensPage++
			}
		}

		// incomplete page indicates there are no more users
		if len(users) < pageSize {
			break
		}
		usersPage++
	}
}

func giteaCheckAndCleanUserToken(client *gitea.Client, cfg *AppConfig, user *gitea.User, token *gitea.AccessToken) {
	// filter tokens by prefix
	if !strings.HasPrefix(token.Name, cfg.GiteaDroneTokenPrefix+"_") {
		return
	}

	// build token is expected to have the following syntax: "<prefix>_<build-id>_<timestamp>"
	nameParts := strings.Split(strings.TrimPrefix(token.Name, cfg.GiteaDroneTokenPrefix+"_"), "_")
	if len(nameParts) != 2 {
		return
	}

	//buildId, err := strconv.Atoi(nameParts[0])
	//if err != nil {
	//	logrus.Errorf("Failed to parse token build-id: user=%s token=%s", user.UserName, token.Name)
	//}
	timestamp, err := strconv.Atoi(nameParts[1])
	if err != nil {
		logrus.Errorf("Failed to parse token timestamp: user=%s token=%s", user.UserName, token.Name)
		return
	}

	now := time.Now().Unix()
	diff := now - int64(timestamp)
	if diff > int64(cfg.GiteaDroneTokenTTL) {
		_, err := client.DeleteAccessToken(token.ID)
		if err != nil {
			logrus.Errorf("Failed to delete token: user=%s token=%s", user.UserName, token.Name)
		}
	}
}
