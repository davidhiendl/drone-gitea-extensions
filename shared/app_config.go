package shared

type AppConfig struct {
	Bind   string `envconfig:"DRONE_BIND"`
	Debug  bool   `envconfig:"DRONE_DEBUG"`
	Secret string `envconfig:"DRONE_SECRET" required:"true"`

	GiteaURL  string `envconfig:"GITEA_URL" required:"true"`
	GiteaUser string `envconfig:"GITEA_USERNAME" required:"true"`
	GiteaPass string `envconfig:"GITEA_PASSWORD" required:"true"`
	// GiteaToken         string `envconfig:"GITEA_TOKEN" required:"true"`

	GiteaDroneTokenTTL      int    `envconfig:"GITEA_DRONE_TOKEN_TTL" default:"3900"`
	GiteaDroneTokenGCEnable bool   `envconfig:"GITEA_DRONE_TOKEN_GC_ENABLE" default:"true" required:"true"`
	GiteaDroneTokenPrefix   string `envconfig:"GITEA_DRONE_TOKEN_PREFIX" default:"drone"`

	DroneConfigIncludeMax      int  `envconfig:"DRONE_CONFIG_INCLUDE_MAX" default:"20"`
	EmulateCIPrefixedVariables bool `envconfig:"EMULATE_CI_PREFIXED_ENV_VARS" default:"true"`
}
