#!/usr/bin/env bash
# Deploy ez para prod (consale)
# Uso: ./scripts/deploy.sh

set -euo pipefail

REMOTE_HOST="consale"
REMOTE_PATH="~/app"
SERVICE_NAME="ez"
APP_NAME="ez"

echo ">> Build local (tailwind + templ + binário linux/amd64)..."
make tailwind-build
make templ-generate
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Environment=production" -o ./bin/"$APP_NAME" ./cmd/main.go

echo ">> Copiando binário..."
scp ./bin/"$APP_NAME" "$REMOTE_HOST":"$REMOTE_PATH"/"$APP_NAME".new

echo ">> Copiando static/..."
rsync -az --delete ./static/ "$REMOTE_HOST":"$REMOTE_PATH"/static/

echo ">> Ativando novo binário e reiniciando serviço..."
ssh "$REMOTE_HOST" "set -e; \
  mv $REMOTE_PATH/$APP_NAME.new $REMOTE_PATH/$APP_NAME && \
  chmod +x $REMOTE_PATH/$APP_NAME && \
  systemctl restart $SERVICE_NAME && \
  sleep 1 && \
  systemctl is-active $SERVICE_NAME"

echo ">> Deploy ok."
