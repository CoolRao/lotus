#!/usr/bin/env bash
export MINER_API_INFO=$(cat ~/.lotusstorage/token):$(cat ~/.lotusstorage/api)
nohup ./lotus-worker --worker-repo=~/.lotusworker01 run --listen=127.0.0.1:34567 --precommit1=true --precommit2=false --commit=false >worker01.log 2>&1 &
