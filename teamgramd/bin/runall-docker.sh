#!/usr/bin/env bash

# Per-service GOMEMLIMIT: total ~1000MiB for 1280m container (280m for OS/cache)
# Services that handle file I/O (dfs, media, bff) get larger budgets.
# GOMEMLIMIT is inherited from environment if not overridden per-service.

echo "run idgen ..."
GOMEMLIMIT=50MiB nohup ./idgen -f=../etc2/idgen.yaml >> ../logs/idgen.log  2>&1 &
sleep 1

echo "run status ..."
GOMEMLIMIT=50MiB nohup ./status -f=../etc2/status.yaml >> ../logs/status.log  2>&1 &
sleep 1

echo "run authsession ..."
GOMEMLIMIT=80MiB nohup ./authsession -f=../etc2/authsession.yaml >> ../logs/authsession.log  2>&1 &
sleep 1

echo "run dfs ..."
GOMEMLIMIT=200MiB nohup ./dfs -f=../etc2/dfs.yaml >> ../logs/dfs.log  2>&1 &
sleep 1

echo "run media ..."
GOMEMLIMIT=100MiB nohup ./media -f=../etc2/media.yaml >> ../logs/media.log  2>&1 &
sleep 1

echo "run biz ..."
GOMEMLIMIT=80MiB nohup ./biz -f=../etc2/biz.yaml >> ../logs/biz.log  2>&1 &
sleep 1

echo "run msg ..."
GOMEMLIMIT=80MiB nohup ./msg -f=../etc2/msg.yaml >> ../logs/msg.log  2>&1 &
sleep 1

echo "run sync ..."
GOMEMLIMIT=80MiB nohup ./sync -f=../etc2/sync.yaml >> ../logs/sync.log  2>&1 &
sleep 1

echo "run bff ..."
GOMEMLIMIT=150MiB nohup ./bff -f=../etc2/bff.yaml >> ../logs/bff.log  2>&1 &
sleep 5

echo "run session ..."
GOMEMLIMIT=100MiB nohup ./session -f=../etc2/session.yaml >> ../logs/session.log  2>&1 &
sleep 1

echo "run gateway ..."
GOMEMLIMIT=100MiB nohup ./gateway -f=../etc2/gateway.yaml >> ../logs/gateway.log  2>&1 &
sleep 1
