#!/bin/bash
set -euo pipefail

VPS_USER="sadmin"
VPS_HOST="${APC_VPS_HOST:?APC_VPS_HOST not set}"
SSH_KEY="${APC_SSH_KEY:-$HOME/.ssh/id_apc}"
REMOTE_DIR="containers/apc"
CADDY_DIR="containers/caddy"

BACKEND_IMG="apc-backend"
ADMIN_IMG="apc-admin"
MEMBER_IMG="apc-member"

# ---------------------------------------------------------------------------
# 1. Start ssh-agent, add key once, kill agent on exit
# ---------------------------------------------------------------------------
echo "-> Starting ssh-agent and adding key (passphrase asked once)..."
eval "$(ssh-agent -s)"
trap 'ssh-agent -k >/dev/null 2>&1' EXIT
ssh-add "$SSH_KEY"

# SSH_AUTH_SOCK is now set in this shell.  Build helpers AFTER the agent is
# ready so they inherit the correct socket path via the environment.
#
# -i "$SSH_KEY"          : tells ssh which identity to offer (required with
#                          IdentitiesOnly so the key is found in the agent)
# -o IdentitiesOnly=yes  : prevents ssh from trying every key in the agent or
#                          the default ~/.ssh/id_* files
# -o StrictHostKeyChecking=no : non-interactive deploy, no known_hosts prompt
SSH_OPTS="-i $SSH_KEY -o IdentitiesOnly=yes -o StrictHostKeyChecking=no"
SSH="ssh $SSH_OPTS"
SCP="scp $SSH_OPTS"

# ---------------------------------------------------------------------------
# Sanity checks
# ---------------------------------------------------------------------------
if [ ! -f .env ]; then
  echo "ERROR: .env missing locally. Seed it from .env.example." >&2
  exit 1
fi

IMAGE_TAG="$(git rev-parse --short HEAD)"
export IMAGE_TAG
echo "-> Deploying tag: $IMAGE_TAG"

# ---------------------------------------------------------------------------
# 2. Build images locally (no DOCKER_HOST — local daemon)
# ---------------------------------------------------------------------------
echo "-> Building images locally..."
docker compose -f docker-compose.prod.yml build --no-cache

# ---------------------------------------------------------------------------
# 3. Tag images locally
# ---------------------------------------------------------------------------
docker tag "$BACKEND_IMG:$IMAGE_TAG" "$BACKEND_IMG:latest"
docker tag "$ADMIN_IMG:$IMAGE_TAG"   "$ADMIN_IMG:latest"
docker tag "$MEMBER_IMG:$IMAGE_TAG"  "$MEMBER_IMG:latest"

# ---------------------------------------------------------------------------
# 4. Transfer images via pipe — agent socket is inherited by ssh subprocess
# ---------------------------------------------------------------------------
echo "-> Transferring images to VPS (this may take a while)..."
docker save "$BACKEND_IMG:$IMAGE_TAG" "$ADMIN_IMG:$IMAGE_TAG" "$MEMBER_IMG:$IMAGE_TAG" \
  | gzip \
  | $SSH "$VPS_USER@$VPS_HOST" "gunzip | docker load"

# ---------------------------------------------------------------------------
# 5. Create remote directories
#    No /var/log/caddy — sadmin has no sudo; create that manually once.
# ---------------------------------------------------------------------------
echo "-> Ensuring remote directories exist..."
$SSH "$VPS_USER@$VPS_HOST" "mkdir -p ~/$REMOTE_DIR/data ~/$CADDY_DIR/sites"

# ---------------------------------------------------------------------------
# 6. SCP files to VPS
#    scp remote paths: ~/ expands on the REMOTE side inside a quoted arg that
#    is passed verbatim, but scp interprets ~ relative to the remote user.
#    Using the bare ~ in the destination is reliable for scp.
# ---------------------------------------------------------------------------
echo "-> Syncing compose + env to VPS..."
$SCP docker-compose.prod.yml "$VPS_USER@$VPS_HOST:~/$REMOTE_DIR/docker-compose.yml"
$SCP .env                    "$VPS_USER@$VPS_HOST:~/$REMOTE_DIR/.env"

echo "-> Syncing Caddy config to VPS..."
$SCP caddy/docker-compose.yml "$VPS_USER@$VPS_HOST:~/$CADDY_DIR/docker-compose.yml"
$SCP caddy/Caddyfile           "$VPS_USER@$VPS_HOST:~/$CADDY_DIR/Caddyfile"
$SCP caddy/sites/admin.conf    "$VPS_USER@$VPS_HOST:~/$CADDY_DIR/sites/admin.conf"
$SCP caddy/sites/member.conf   "$VPS_USER@$VPS_HOST:~/$CADDY_DIR/sites/member.conf"

# ---------------------------------------------------------------------------
# 7. Start Caddy first (creates the caddy-proxy network)
# ---------------------------------------------------------------------------
echo "-> Starting Caddy stack on VPS..."
$SSH "$VPS_USER@$VPS_HOST" "
  cd ~/$CADDY_DIR
  docker compose up -d
  docker exec caddy caddy reload --config /etc/caddy/Caddyfile 2>/dev/null || true
"

# ---------------------------------------------------------------------------
# 8. Start app stack on VPS (caddy-proxy network now exists)
# ---------------------------------------------------------------------------
echo "-> Starting app stack on VPS..."
$SSH "$VPS_USER@$VPS_HOST" "
  cd ~/$REMOTE_DIR
  docker compose up -d
"

echo "-> Done ($IMAGE_TAG)"
