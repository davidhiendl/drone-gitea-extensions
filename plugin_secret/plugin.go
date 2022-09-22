package plugin_secret

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-secret-extension/shared"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/secret"
	"github.com/sirupsen/logrus"
	"net/url"
)

const PATH_GITEA = "gitea"
const KEY_GITEA_URL = "url"                         // gitea.example.com
const KEY_GITEA_BUILD_TOKEN = "build_token"         // secret
const KEY_GITEA_PACKAGES_API = "packages_url"       // https://gitea.example.com/api/packages
const KEY_GITEA_DOCKER_REGISTRY = "docker_registry" // gitea.example.com

// New returns a new secret plugin.
func New(client *gitea.Client, config *shared.AppConfig, cache *shared.TokenCache) secret.Plugin {

	giteaUrl, err := url.Parse(config.GiteaURL)
	if err != nil {
		logrus.Fatalf("failed to parse gitea url: %+v", err)
	}

	return &plugin{
		client:                  client,
		config:                  config,
		cache:                   cache,
		giteaPackagesURL:        config.GiteaURL + "/api/packages",
		giteaURL:                config.GiteaURL,
		giteaDockerRegistryHost: giteaUrl.Hostname(),
	}
}

type plugin struct {
	client                  *gitea.Client
	config                  *shared.AppConfig
	cache                   *shared.TokenCache
	giteaPackagesURL        string
	giteaURL                string
	giteaDockerRegistryHost string
}

func (p *plugin) Find(ctx context.Context, req *secret.Request) (*drone.Secret, error) {
	logrus.Debugf("secret plugin request received: path=%s name=%sbuild=%+v repo=%+v", req.Path, req.Name, req.Build, req.Repo)

	// only handle request for gitea path
	if req.Path != PATH_GITEA {
		return nil, nil
	}

	if req.Name == KEY_GITEA_URL {
		return &drone.Secret{
			Name: KEY_GITEA_URL,
			Data: p.giteaURL,
		}, nil
	}

	if req.Name == KEY_GITEA_PACKAGES_API {
		return &drone.Secret{
			Name: KEY_GITEA_PACKAGES_API,
			Data: p.giteaPackagesURL,
		}, nil
	}

	if req.Name == KEY_GITEA_DOCKER_REGISTRY {
		return &drone.Secret{
			Name: KEY_GITEA_DOCKER_REGISTRY,
			Data: p.giteaDockerRegistryHost,
		}, nil
	}

	if req.Name == KEY_GITEA_BUILD_TOKEN {
		token, err := p.cache.GetAccessToken(req.Build.ID, req.Build.Sender)
		if err != nil {
			logrus.Errorf("%+v", err)
			return nil, err
		}

		return &drone.Secret{
			Name:        KEY_GITEA_BUILD_TOKEN,
			Data:        token.Token,
			PullRequest: false, // never expose to pulls requests
		}, nil
	}

	return nil, nil
}
