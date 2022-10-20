package plugin_registry

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-extensions/shared"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/registry"
	"github.com/sirupsen/logrus"
	"net/url"
)

// New returns a new secret plugin.
func New(client *gitea.Client, config *shared.AppConfig, cache *shared.TokenCache) registry.Plugin {

	giteaUrl, err := url.Parse(config.GiteaURL)
	if err != nil {
		logrus.Fatalf("failed to parse gitea url: %+v", err)
	}

	return &plugin{
		client:                  client,
		config:                  config,
		cache:                   cache,
		giteaDockerRegistryHost: giteaUrl.Hostname(),
	}
}

type plugin struct {
	client                  *gitea.Client
	config                  *shared.AppConfig
	cache                   *shared.TokenCache
	giteaDockerRegistryHost string
}

func (p *plugin) List(ctx context.Context, req *registry.Request) ([]*drone.Registry, error) {
	logrus.Infof("[registry] request for build=%s %s/%s commit=%s", req.Build.ID, req.Repo.Namespace, req.Repo.Name, req.Build.After)

	logrus.Debugf("registry plugin request received: build=%+v repo=%+v", req.Build, req.Repo)

	token, err := p.cache.GetAccessToken(req.Build.ID, req.Build.Sender)
	if err != nil {
		logrus.Errorf("%+v", err)
		return nil, err
	}

	credentials := []*drone.Registry{
		{
			Address:  p.giteaDockerRegistryHost,
			Username: req.Build.Sender,
			Password: token.Token,
		},
	}

	return credentials, nil
}
