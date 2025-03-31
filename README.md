# Breakglass
Golang application that allows for short-term elevation of privileges in an emergency situation.

## Overview

## Configuration
App should be configured using config.yaml:
```yaml
server:
  listenAddress: :8080
  tlsCertFile: /some/file.crt # optional
  tlsKeyFile: /some/file.key # optional for https
authorizationserver:
  url: http://127.0.0.1:8080
  jwksEndpoint: "realms/master/protocol/openid-connect/certs" # sample for keycloak
frontend:
  oidcAuthority: http://127.0.0.1:8080/realms/master
  oidcClientID: breakglass-ui
  baseURL: http://localhost:8080
mail:
  host: 127.0.0.1
  port: 1025
  insecureSkipVerify: false
kubernetes:
  context: "" # kubectl config context if empty default will be used

```
See `config.example.yaml` for reference.

## Building Docker image
To build docker image:
```bash
docker build -t breakglass .
```
## Kubernetes Deployment
To deploy `CRD`, `RBAC` and application as `Deployment` configure `./config/default/config.yaml` with proper cluster
related configuration and run:
```bash
make deploy
```

Make sure that `${authorizationserver.url}/${jwksEndpoint}` is accessible by application. <br/>
See `config/dev/resources/keycloak.yaml` for sample kubernetes `Service` exposing keycloak as auth server.

### Dev environment:
```bash
make deploy_dev
```
Will perform similar deployment as standard `deploy`, but will additionally include deployment of keycloak and mailhog alongside
with services pointing to their open ports. It also includes NodePort type services so that application can be instantly
accessed.

#### Accessing app through docker kind cluster
It was tested for `kind` docker single cluster. <br/>
Assuming your docker container has ip of `172.19.0.2` add following entry to `/etc/hosts`: `172.19.0.2      breakglass-dev`.
Then you should be able to access main breakglass app under: `https://breakglass-dev:30081`, keycloak under: `https://breakglass-dev:30083` 
and mailhog under `http://breakglass-dev:30084`.

#### First time configuring Keycloak
Go to `Clients` tab -> Create -> Add client called `breakglass-ui` or same as `breakglass-config.oidcClientID` -> Set
correct `Valid Redirect URIs` and `Web Origins` (for testing and developement setting all `*` will work).
