#!/usr/bin/env bash
export MINER_API_INFO=$(cat ~/.lotusstorage/token):$(cat ~/.lotusstorage/api)
nohup ./lotus-worker --worker-repo=~/.lotusworker02 run --listen=127.0.0.1:34568 --precommit1=false --precommit2=true --commit=false >worker02.log 2>&1 &
