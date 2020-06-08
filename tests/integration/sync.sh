#!/usr/bin/env bash

RED='\033[0;31m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color


for account in $(jq -r ".accounts[] | [.xpub,.derivationMode] | @csv" < ~/.sats.json | sed "s/\"//g")
do
    xpub=$(echo "$account" | cut -d"," -f1)
    scheme=$(echo "$account" | cut -d"," -f2)

    echo -e ""
    echo -e "${CYAN}SCHEME:${NC} ${scheme}"
    echo -e "${CYAN}XPUB:${NC}   ${xpub}"

    cmd="time ledger-live sync --xpub $xpub -c bitcoin_testnet -s ${scheme/standard/} -f summary"
    echo -e "${CYAN}SYNC:${NC}   explorers.api.live.ledger.com"
    want=$($cmd)
    echo -e "${CYAN}SYNC:${NC}   0.0.0.0:20000"
    got=$(EXPERIMENTAL_EXPLORERS=1 EXPLORER="http://0.0.0.0:20000" $cmd)

    if [ "$got" = "$want" ]; then
      echo -e "${CYAN}OUTPUT:${NC} $got"
      echo -e "${GREEN}Synchronization OK ✅${NC}"
    else
        echo -e "${RED}WANT:${NC}    $want"
        echo -e "${RED}GOT:${NC}     $got"
        echo -e "${RED}Unexpected response ❌${NC}"
        exit 1
    fi
done
