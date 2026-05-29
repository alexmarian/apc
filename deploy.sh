#!/bin/bash
set -euo pipefail

VPS_USER="apcadmin"
VPS_HOST="${APC_VPS_HOST:?APC_VPS_HOST not set}"
SSH_KEY="${APC_SSH_KEY:-$HOME/.ssh/id_apcadmin}"
REMOTE_DIR="\$HOME/containers/apc"

SCP="scp -i $SSH_KEY"
BACKEND_IMG="apc-backend"
ADMIN_IMG="apc-admin"
MEMBER_IMG="apc-member"

cleanup() {
  [ -n "${SSH_AGENT_PID:-}" ] && ssh-agent -k >/dev/null 2>&1 || true
}
trap cleanup EXIT

if [ ! -f .env ]; then
  echo "✗ .env missing locally. Seed it from .env.example." >&2
  exit 1
fi

IMAGE_TAG="$(git rev-parse --short HEAD)"
export IMAGE_TAG
echo "→ Deploying tag: $IMAGE_TAG"

echo "→ Adding key to ssh-agent..."
eval "$(ssh-agent)" >/dev/null
ssh-add "$SSH_KEY"

docker context inspect vps >/dev/null 2>&1 || \
  docker context create vps --docker "host=ssh://$VPS_USER@$VPS_HOST"

echo "→ Building images locally..."
docker compose -f docker-compose.prod.yml build
docker tag "$BACKEND_IMG:latest" "$BACKEND_IMG:$IMAGE_TAG"
docker tag "$ADMIN_IMG:latest"   "$ADMIN_IMG:$IMAGE_TAG"
docker tag "$MEMBER_IMG:latest"  "$MEMBER_IMG:$IMAGE_TAG"

echo "→ Transferring images to VPS..."
docker save "$BACKEND_IMG:$IMAGE_TAG" "$ADMIN_IMG:$IMAGE_TAG" "$MEMBER_IMG:$IMAGE_TAG" \
  | gzip \
  | ssh -i "$SSH_KEY" "$VPS_USER@$VPS_HOST" "gunzip | docker load"

echo "→ Tagging :latest on VPS..."
docker --context vps tag "$BACKEND_IMG:$IMAGE_TAG" "$BACKEND_IMG:latest"
docker --context vps tag "$ADMIN_IMG:$IMAGE_TAG"   "$ADMIN_IMG:latest"
docker --context vps tag "$MEMBER_IMG:$IMAGE_TAG"  "$MEMBER_IMG:latest"

echo "→ Syncing compose + env to VPS..."
$SCP docker-compose.prod.yml "$VPS_USER@$VPS_HOST:$REMOTE_DIR/docker-compose.yml"
$SCP .env                    "$VPS_USER@$VPS_HOST:$REMOTE_DIR/.env"

echo "→ Starting on VPS..."
docker --context vps compose -f "$REMOTE_DIR/docker-compose.yml" up -d

echo "✓ Done ($IMAGE_TAG)"
