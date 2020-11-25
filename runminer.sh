#!/usr/bin/env bash

sed -i '/AllowAddPiece = /c\  AllowAddPiece = false'  ~/.lotusstorage/config.toml
sed -i '/AllowPreCommit1 = /c\  AllowPreCommit1 = false'  ~/.lotusstorage/config.toml
sed -i '/AllowPreCommit2 = /c\  AllowPreCommit2 = false'  ~/.lotusstorage/config.toml
sed -i '/AllowCommit = /c\  AllowCommit = false'  ~/.lotusstorage/config.toml

nohup ./lotus-miner run --nosync >miner.log 2>&1 &
