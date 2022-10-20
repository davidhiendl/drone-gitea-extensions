package plugin_config

import (
	"code.gitea.io/sdk/gitea"
	"context"
	"dhswt.de/drone-gitea-extensions/shared"
	"errors"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/config"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// New returns a new secret Plugin.
func New(client *gitea.Client, config *shared.AppConfig, cache *shared.TokenCache) *Plugin {

	giteaUrl, err := url.Parse(config.GiteaURL)
	if err != nil {
		logrus.Fatalf("failed to parse gitea url: %+v", err)
	}

	return &Plugin{
		client:                  client,
		config:                  config,
		cache:                   cache,
		giteaPackagesURL:        config.GiteaURL + "/api/packages",
		giteaURL:                config.GiteaURL,
		giteaDockerRegistryHost: giteaUrl.Hostname(),
	}
}

type Plugin struct {
	client                  *gitea.Client
	config                  *shared.AppConfig
	cache                   *shared.TokenCache
	giteaPackagesURL        string
	giteaURL                string
	giteaDockerRegistryHost string
}

func (p *Plugin) Find(ctx context.Context, req *config.Request) (*drone.Config, error) {
	contents, resp, err := p.client.GetContents(req.Repo.Namespace, req.Repo.Name, req.Build.After, ".drone.yml")
	if err != nil {
		return nil, err
	}
	// handle statusCode != 200
	if resp.StatusCode != 200 {
		return nil, errors.New("Failed to find .drone.yml in gitea, status=" + resp.Status)
	}
	// handle invalid content type
	if contents.Type != "file" {
		return nil, errors.New("Failed to find .drone.yml in gitea, type=" + contents.Type + " expected: file")
	}

	originalYaml := contents.Content

	yaml, err := p.regexReplaceIncludeDirectives(*originalYaml, http.DefaultClient)
	if err != nil {
		return nil, err
	}

	return &drone.Config{
		Data: yaml,
	}, nil
}

var directiveIncludeRegex = regexp.MustCompile(`(?:^|\n)_include\s*:\s*["']?(.*)["']?\s*`)

func (p *Plugin) regexReplaceIncludeDirectives(yaml string, httpClient *http.Client) (string, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	includesApplied := 0

	for true {
		if includesApplied >= p.config.DroneConfigIncludeMax {
			return "", errors.New("to many include directives, max=" + strconv.Itoa(p.config.DroneConfigIncludeMax))
		}

		match := directiveIncludeRegex.FindStringIndex(yaml)
		if match == nil {
			return yaml, nil
		}

		directive := yaml[match[0]:match[1]]
		directiveValueMatch := directiveIncludeRegex.FindStringSubmatch(directive)
		if directiveValueMatch != nil && len(directiveValueMatch) == 2 {
			directiveValue := directiveValueMatch[1]
			directiveYaml, err := p.getUrlBodyAsStr(httpClient, directiveValue)
			if err != nil {
				return "", err
			}

			directiveCommentStart := "# DIRECTIVE_START " + strings.Trim(directive, "\n ") + "\n"
			directiveCommentEnd := "# DIRECTIVE_END " + strings.Trim(directive, "\n ") + "\n"

			// splice yaml into string and continue loop
			yaml = yaml[0:match[0]] +
				"\n" + directiveCommentStart +
				strings.Trim(directiveYaml, "\n ") +
				"\n" + directiveCommentEnd + yaml[match[1]:]

			includesApplied++
		} else {
			return "", errors.New("failed to extract value from include directive")
		}
	}

	return "", errors.New("failed to process include directive")
}

func (p *Plugin) getUrlBodyAsStr(httpClient *http.Client, url string) (string, error) {
	// TODO detect gitea url and fetch resource using authentication if needed

	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("status code for " + url + " != 200")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
