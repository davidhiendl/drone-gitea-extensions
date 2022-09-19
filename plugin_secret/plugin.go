package plugin_secret

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-secret-extension/shared"
	"errors"
	"fmt"
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
func New(client *gitea.Client, config *shared.AppConfig) secret.Plugin {

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

func (p *plugin) Find(ctx context.Context, req *secret.Request) (*drone.Secret, error) {

	logrus.Infof("plugin request received: path=%s name=%s", req.Path, req.Name)
	logrus.Infof("> build=%+v", req.Build)
	logrus.Infof("> repo=%+v", req.Repo)

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
		token, err := p.createGiteaToken(req)
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

func (p *plugin) createGiteaToken(req *secret.Request) (*gitea.AccessToken, error) {
	if len(req.Build.Sender) == 0 {
		return nil, errors.New(fmt.Sprintf("build is missing sender info: path=%s name=%s repo=%+v build=%+v", req.Path, req.Name, req.Repo, req.Build))
	}

	p.client.SetSudo("")
	_, _, err := p.client.GetUserInfo(req.Build.Sender)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to find gitea user: '%s'", req.Build.Sender))
	}

	accessTokenName := fmt.Sprintf("%s_%s_%s", p.config.GiteaDroneTokenPrefix, req.Build.ID, req.Build.Timestamp)

	p.client.SetSudo(req.Build.Sender)
	token, _, err := p.client.CreateAccessToken(gitea.CreateAccessTokenOption{Name: accessTokenName})
	if err != nil {
		// TODO handle already exists errors?
		return nil, errors.New(fmt.Sprintf("failed to create gitea access token: err=%+v path=%s name=%s repo=%+v build=%+v", err, req.Path, req.Name, req.Repo, req.Build))
	}

	return token, nil
}
