#!/bin/sh
# Update production to the latest pushed main: pull, build the engine and the
# static frontend as the unprivileged `lofp` user, then restart the service.
# Run on the production host as a sudo-capable user: sh /srv/lofp/app/deploy/prod-update.sh
set -eu

APP=/srv/lofp/app
RUN="sudo -u lofp env HOME=/srv/lofp PATH=/usr/local/go/bin:/opt/node/bin:/usr/bin:/bin"

$RUN git -C "$APP" pull --ff-only
$RUN git -C "$APP" log --oneline -1

# Engine
$RUN sh -c "cd $APP/engine && CGO_ENABLED=0 go build -o lofp.new ./cmd/lofp"

# Frontend: static build served by the engine (LOFP_STATIC_DIR). Never run the
# Vite dev server in production. --ignore-scripts keeps npm package install
# hooks from running on this host. VITE_* build variables, if any, come from
# /srv/lofp/frontend.env.
$RUN sh -c "cd $APP/frontend && npm ci --ignore-scripts --no-audit --no-fund && \
  { [ -f /srv/lofp/frontend.env ] && set -a && . /srv/lofp/frontend.env && set +a; true; } && \
  npm run build"

$RUN mv "$APP/engine/lofp.new" "$APP/engine/lofp"
sudo systemctl restart lofp
sleep 5
systemctl is-active lofp
curl -fsS -o /dev/null -w 'healthz %{http_code}\n' http://localhost:8080/healthz
