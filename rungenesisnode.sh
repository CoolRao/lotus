#!/usr/bin/env bash
nohup ./lotus daemon --lotus-make-genesis=devgen.car --genesis-template=localnet.json --bootstrap=false >lotus.log 2>&1 &