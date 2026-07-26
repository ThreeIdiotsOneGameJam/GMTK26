# Server deployment

The dedicated server is published to:

```text
ghcr.io/threeidiotsonegamejam/gmtk26
```

It accepts WebSocket connections on TCP port `58008` by default. Active games
are held in memory, so restarting the container ends any games in progress.

## Publish an image

1. Open **Actions** in the GitHub repository.
2. Select **Build and publish server image**.
3. Select **Run workflow**, choose the commit or branch, and enter an image
   tag. The default is `latest`.

The workflow runs the Docker build on a GitHub-hosted runner and publishes two
GHCR tags:

- The tag entered when starting the workflow, such as `latest` or `v1.0.0`.
- An immutable `sha-<full commit SHA>` tag for rollback.

The workflow authenticates with its short-lived `GITHUB_TOKEN`; no registry
password needs to be added to the repository. The first published package may
be private. In the package settings, either make it public or grant the
deployment server an account that can read it.

## Prepare the host

Install Docker Engine with the Compose plugin on a Linux server, then copy
[`docker-compose.yml`](../docker-compose.yml) into a deployment directory:

```sh
sudo install -d -m 0755 /opt/gmtk26
sudo cp docker-compose.yml /opt/gmtk26/docker-compose.yml
cd /opt/gmtk26
```

Create `/opt/gmtk26/.env`:

```dotenv
IMAGE_TAG=latest
SERVER_BIND_IP=0.0.0.0
SERVER_PORT=58008
```

- `IMAGE_TAG` selects the published GHCR tag.
- `SERVER_BIND_IP` controls which host interface publishes the port. Keep
  `0.0.0.0` for direct public access, or use `127.0.0.1` when a reverse proxy
  on the same host provides TLS.
- `SERVER_PORT` changes both the host and container port.

If the GHCR package is private, create a GitHub personal access token (classic)
with `read:packages`, then authenticate without placing the token in the
Compose file:

```sh
read -rsp "GHCR token: " GHCR_TOKEN
printf '%s' "$GHCR_TOKEN" | docker login ghcr.io \
  --username YOUR_GITHUB_USERNAME \
  --password-stdin
unset GHCR_TOKEN
```

Public packages do not require `docker login`.

## Start the server

Pull the selected image and start it:

```sh
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100 server
```

For direct public access, allow the configured TCP port in both the host
firewall and the hosting provider's network firewall. For example, on a host
using UFW:

```sh
sudo ufw allow 58008/tcp
```

Clients can then connect to `ws://SERVER_ADDRESS:58008`.

## TLS and secure WebSockets

An HTTPS-hosted browser client must connect over `wss://`, not `ws://`. Set
`SERVER_BIND_IP=127.0.0.1` and place a TLS-terminating reverse proxy in front
of the server. Build the client with `ClientWebSocketScheme` in
`src/constants/constants.go` set to `wss`. For example, a Caddy site can proxy
the WebSocket endpoint:

```caddyfile
game.example.com {
	reverse_proxy 127.0.0.1:58008
}
```

Allow TCP ports `80` and `443` through the firewalls, and do not expose
`58008` publicly in this setup. Clients connect to
`wss://game.example.com`.

## Upgrade or roll back

To deploy a newly published `latest` image:

```sh
docker compose pull
docker compose up -d
```

For a pinned release or rollback, change `IMAGE_TAG` in `.env` to a published
version or `sha-<full commit SHA>`, then run:

```sh
docker compose pull
docker compose up -d
```

Check the running image and follow logs with:

```sh
docker compose images
docker compose logs -f server
```

Stop the deployment with `docker compose down`.
