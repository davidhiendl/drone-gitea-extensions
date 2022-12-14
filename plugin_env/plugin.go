package plugin_env

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-extensions/shared"
	"github.com/Masterminds/semver"
	"github.com/drone/drone-go/plugin/environ"
	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
	"strings"
)

// New returns a new secret plugin.
func New(client *gitea.Client, config *shared.AppConfig, cache *shared.TokenCache) environ.Plugin {

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
		giteaHostname:           giteaUrl.Hostname(),
	}
}

type plugin struct {
	client                  *gitea.Client
	config                  *shared.AppConfig
	cache                   *shared.TokenCache
	giteaPackagesURL        string
	giteaURL                string
	giteaDockerRegistryHost string
	giteaHostname           string
}

func (p *plugin) List(ctx context.Context, req *environ.Request) ([]*environ.Variable, error) {
	logrus.Infof("[env] request for build=%d %s/%s commit=%s", req.Build.ID, req.Repo.Namespace, req.Repo.Name, req.Build.After)

	logrus.Debugf("environment plugin request received: build=%+v repo=%+v", req.Build, req.Repo)

	token, err := p.cache.GetAccessToken(req.Build.ID, req.Build.Sender)
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

	if p.config.EmulateCIPrefixedVariables {
		ciVariables := []*environ.Variable{
			// mirror various gitlab CI_ variables
			{Name: "CI_SERVER_URL", Data: p.giteaURL, Mask: false},
			{Name: "CI_SERVER_HOST", Data: p.giteaHostname, Mask: false},
			{Name: "CI_PROJECT_NAME", Data: req.Repo.Name, Mask: false},
			{Name: "CI_PROJECT_NAMESPACE", Data: req.Repo.Namespace, Mask: false},

			{Name: "CI_REGISTRY", Data: p.giteaDockerRegistryHost, Mask: false},
			{Name: "CI_REGISTRY_IMAGE", Data: p.giteaDockerRegistryHost + "/" + req.Repo.Namespace + "/" + req.Repo.Name, Mask: false},
			{Name: "CI_REGISTRY_USER", Data: token.Name, Mask: false},
			{Name: "CI_REGISTRY_PASSWORD", Data: token.Token, Mask: true},
		}
		envVars = append(envVars, ciVariables...)
	}

	if p.config.EnvAddTagSemver && strings.HasPrefix(req.Build.Ref, "refs/tags/") {
		tag := strings.TrimPrefix(req.Build.Ref, "refs/tags/")
		v, err := semver.NewVersion(tag)
		if err != nil {
			logrus.Debugf("failed to ref as semver: %s", req.Build.Ref)
		}

		semverVars := []*environ.Variable{
			// mirror various gitlab CI_ variables
			{Name: "SEMVER_MAJOR", Data: strconv.FormatInt(v.Major(), 10), Mask: false},
			{Name: "SEMVER_MINOR", Data: strconv.FormatInt(v.Minor(), 10), Mask: false},
			{Name: "SEMVER_PATCH", Data: strconv.FormatInt(v.Patch(), 10), Mask: false},
			{Name: "SEMVER_PRERELEASE", Data: v.Prerelease(), Mask: false},
			{Name: "SEMVER_METADATA", Data: v.Metadata(), Mask: false},
			{
				Name: "SEMVER_MAJOR_MINOR",
				Data: strconv.FormatInt(v.Major(), 10) +
					"." + strconv.FormatInt(v.Minor(), 10),
				Mask: false,
			},
			{
				Name: "SEMVER_MAJOR_MINOR_PATCH",
				Data: strconv.FormatInt(v.Major(), 10) +
					"." + strconv.FormatInt(v.Minor(), 10) +
					"." + strconv.FormatInt(v.Patch(), 10),
				Mask: false,
			},
		}
		envVars = append(envVars, semverVars...)
	}

	return envVars, nil
}
