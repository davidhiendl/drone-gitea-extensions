package main

import (
	"dhswt.de/drone-gitea-secret-extension/shared"
	"net/http"

	"dhswt.de/drone-gitea-secret-extension/plugin"
	"github.com/drone/drone-go/plugin/secret"

	"code.gitea.io/sdk/gitea"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := new(shared.AppConfig)
	err := envconfig.Process("", cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if cfg.Secret == "" {
		logrus.Fatalln("missing secret key")
	}
	if cfg.Bind == "" {
		cfg.Bind = ":3000"
	}

	client := createGiteaClient(cfg)

	handler := secret.Handler(
		cfg.Secret,
		plugin.New(client, cfg),
		logrus.StandardLogger(),
	)

	logrus.Infof("server listening on address %s", cfg.Bind)

	http.Handle("/", handler)
	logrus.Fatal(http.ListenAndServe(cfg.Bind, nil))
}

func createGiteaClient(cfg *shared.AppConfig) *gitea.Client {
	// gitea.SetToken(cfg.GiteaToken)
	client, err := gitea.NewClient(cfg.GiteaURL, gitea.SetBasicAuth(cfg.GiteaUser, cfg.GiteaPass))
	if err != nil {
		logrus.Fatalf("Failed to create gitea client, %v", err)
	}
	return client
}
