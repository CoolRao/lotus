#!/usr/bin/env bash
pkill -9 lotus
rm -rf ~/.lotus ~/.genesis-sctors localnet.json devgen.car
./lotus-seed pre-seal --sector-size 2KiB --num-sectors 2
./lotus-seed genesis new localnet.json
./lotus-seed genesis add-miner localnet.json ~/.genesis-sectors/pre-seal-t01000.json