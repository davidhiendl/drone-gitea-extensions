package main

import (
	"code.gitea.io/sdk/gitea"
	"dhswt.de/drone-gitea-secret-extension/plugin_env"
	"dhswt.de/drone-gitea-secret-extension/plugin_registry"
	"dhswt.de/drone-gitea-secret-extension/plugin_secret"
	"dhswt.de/drone-gitea-secret-extension/shared"
	"github.com/drone/drone-go/plugin/environ"
	"github.com/drone/drone-go/plugin/registry"
	"github.com/drone/drone-go/plugin/secret"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"net/http"
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
	tokenCache := shared.NewTokenCache(client, cfg)
	tokenCache.StartCleanupAccessTokenJob()

	environHandler := environ.Handler(
		cfg.Secret,
		plugin_env.New(client, cfg, &tokenCache),
		logrus.StandardLogger(),
	)

	secretHandler := secret.Handler(
		cfg.Secret,
		plugin_secret.New(client, cfg, &tokenCache),
		logrus.StandardLogger(),
	)

	registryHandler := registry.Handler(
		cfg.Secret,
		plugin_registry.New(client, cfg, &tokenCache),
		logrus.StandardLogger(),
	)

	switch cfg.DefaultExtension {
	case "environ":
		http.Handle("/", environHandler)
		break
	case "secret":
		http.Handle("/", secretHandler)
		break
	case "registry":
		http.Handle("/", registryHandler)
		break
	default:
		logrus.Fatalf("no valid handler specified for DRONE_DEFAULT_EXTENSION, valid values: 'environ' (default), 'secret', 'registry'")
	}

	http.Handle("/env", environHandler)
	http.Handle("/secret", secretHandler)
	http.Handle("/registry", registryHandler)

	if cfg.GiteaDroneTokenGCEnable {
		shared.StartGiteaTokenCleanupBackgroundJob(client, cfg)
	}

	logrus.Infof("server listening on address %s", cfg.Bind)
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
