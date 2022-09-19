# Drone Gitea Secret Extension

A secret extension for Gitea to create temporary per-pipeline access tokens with access scoped to the owner of the build job.

## Installation

Create a shared secret:

```bash
openssl rand -hex 16
```

Download and run the plugin:

```console
$ docker run -d \
  -p 3000:3000 \
  --env=DRONE_DEBUG=true \
  --env=DRONE_SECRET=<shared-secret> \
  --restart=always \
  --name=drone-gitea-secret-extension
```

Update your runner configuration to include the plugin address and the shared secret.

```bash
DRONE_SECRET_PLUGIN_ENDPOINT=http://1.2.3.4:3000
DRONE_SECRET_PLUGIN_TOKEN=<shared-secret>
```

## Use in pipelines
```yaml
kind: pipeline
name: default

steps:
- name: build
  image: alpine
  environment:
    GITEA_URL:
      from_secret: gitea_url
    GITEA_TOKEN:
      from_secret: gitea_build_token
    GITEA_PACKAGES_URL:
      from_secret: gitea_docker_registry

---
kind: secret
name: gitea_url
get:
  path: gitea
  name: url

---
kind: secret
name: gitea_build_token
get:
  path: gitea
  name: build_token

---
kind: secret
name: gitea_packages_url
get:
  path: gitea
  name: packages_url

---
kind: secret
name: gitea_docker_registry
get:
  path: gitea
  name: docker_registry
```
