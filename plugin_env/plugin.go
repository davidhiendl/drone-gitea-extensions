package plugin_env

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-secret-extension/shared"
	"errors"
	"fmt"
	"github.com/drone/drone-go/plugin/environ"
	"github.com/sirupsen/logrus"
	"net/url"
)

// New returns a new secret plugin.
func New(client *gitea.Client, config *shared.AppConfig) environ.Plugin {

	giteaUrl, err := url.Parse(config.GiteaURL)
	if err != nil {
		logrus.Fatalf("failed to parse gitea url: %+v", err)
	}

	return &plugin{
		client:                  client,
		config:                  config,
		giteaPackagesURL:        config.GiteaURL + "/api/packages",
		giteaURL:                config.GiteaURL,
		giteaDockerRegistryHost: giteaUrl.Hostname(),
	}
}

type plugin struct {
	client                  *gitea.Client
	config                  *shared.AppConfig
	giteaPackagesURL        string
	giteaURL                string
	giteaDockerRegistryHost string
}

func (p *plugin) List(ctx context.Context, req *environ.Request) ([]*environ.Variable, error) {
	logrus.Infof("plugin request received: build=%+v repo=%+v", req.Build, req.Repo)

	token, err := p.createGiteaToken(req)
	if err != nil {
		logrus.Errorf("%+v", err)
		return nil, err
	}

	envVars := []*environ.Variable{
		{Name: "GITEA_URL", Data: p.giteaURL, Mask: false},
		{Name: "GITEA_BUILD_TOKEN", Data: token.Token, Mask: true},
		{Name: "GITEA_PACKAGES_API", Data: p.giteaPackagesURL, Mask: false},
		{Name: "GITEA_DOCKER_REGISTRY", Data: p.giteaDockerRegistryHost, Mask: false},
	}

	return envVars, nil
}

func (p *plugin) createGiteaToken(req *environ.Request) (*gitea.AccessToken, error) {
	if len(req.Build.Sender) == 0 {
		return nil, errors.New(fmt.Sprintf("build is missing sender info: repo=%+v build=%+v", req.Repo, req.Build))
	}

	p.client.SetSudo("")
	_, _, err := p.client.GetUserInfo(req.Build.Sender)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to find gitea user: '%s'", req.Build.Sender))
	}

	accessTokenName := fmt.Sprintf("%s_%d_%d", p.config.GiteaDroneTokenPrefix, req.Build.ID, req.Build.Timestamp)

	p.client.SetSudo(req.Build.Sender)
	token, _, err := p.client.CreateAccessToken(gitea.CreateAccessTokenOption{Name: accessTokenName})
	if err != nil {
		// TODO handle already exists errors?
		return nil, errors.New(fmt.Sprintf("failed to create gitea access token: err=%+v repo=%+v build=%+v", err, req.Repo, req.Build))
	}

	return token, nil
}
