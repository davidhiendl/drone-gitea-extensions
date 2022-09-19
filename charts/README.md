# Deploy via Helm

## Example values.yaml

A minimal values file for helm with only the absolutely required variables.

```yaml
config:
  DRONE_SECRET: "xxxxxxxx"
  GITEA_URL: "https://gitea.example.com"
  GITEA_USERNAME: "example.admin"
  GITEA_PASSWORD: "example"

ingress:
  enabled: true
  hosts:
    - host: gitea-drone-secret-extension.example.com
      paths:
        - path: /
          pathType: ImplementationSpecific
  tls:
    - secretName: ingress-tls-gitea-drone-secret-extension
      hosts:
        - gitea-drone-secret-extension.example.com
```
