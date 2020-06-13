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
    echo -e "${CYAN}SCHEME:${NC}         ${scheme}"
    echo -e "${CYAN}XPUB:${NC}           ${xpub}"

    cmd="time ledger-live sync --xpub $xpub -c bitcoin_testnet -s ${scheme/standard/} -f summary"

    ledger-live libcoreReset
    echo -e "${CYAN}SYNC NOCACHE:${NC}   explorers.api.live.ledger.com"
    wantnc=$($cmd)
    echo -e "${CYAN}SYNC CACHE:${NC}     explorers.api.live.ledger.com"
    _=$($cmd)  # only useful for benchmarks

    ledger-live libcoreReset
    echo -e "${CYAN}SYNC NOCACHE:${NC}   0.0.0.0:20000"
    gotnc=$(EXPERIMENTAL_EXPLORERS=1 EXPLORER="http://0.0.0.0:20000" $cmd)

    echo -e "${CYAN}SYNC CACHE:${NC}     0.0.0.0:20000"
    gotc=$(EXPERIMENTAL_EXPLORERS=1 EXPLORER="http://0.0.0.0:20000" $cmd)

    if [ "$gotnc" = "$wantnc" ]; then
      echo -e "${CYAN}OUTPUT:${NC}         $gotnc"
      echo -e "${GREEN}Synchronization OK ✅${NC}"
    else
        echo -e "${RED}WANT:${NC}           $wantnc"
        echo -e "${RED}GOT:${NC}            $gotnc"
        echo -e "${RED}Unexpected response ❌${NC}"
        exit 1
    fi

    if [ "$gotnc" = "$gotc" ]; then
      echo -e "${GREEN}Cached synchronization OK ✅${NC}"
    else
        echo -e "${RED}WANT:${NC}           $gotnc"
        echo -e "${RED}GOT:${NC}            $gotc"
        echo -e "${RED}Unexpected response ❌${NC}"
        exit 1
    fi
done
