#!/usr/bin/env bash

RED='\033[0;31m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pkh="^pkh\(.*([xt]pub[a-zA-Z0-9]+).*"
wpkh="^wpkh\(.*([xt]pub[a-zA-Z0-9]+).*"
sh_wpkh="^sh\(wpkh\(.*([xt]pub[a-zA-Z0-9]+).*"


for descriptor in $(jq -r ".accounts[] | [.external] | @csv" < lss.json | sed "s/\"//g")
do
    if [[ $descriptor =~ $pkh ]]; then
      xpub="${BASH_REMATCH[1]}"
      scheme=""
    elif [[ $descriptor =~ $sh_wpkh ]]; then
      xpub="${BASH_REMATCH[1]}"
      scheme="segwit"
    elif [[ $descriptor =~ $wpkh ]]; then
      xpub="${BASH_REMATCH[1]}"
      scheme="native_segwit"
    else
      exit 1
    fi

    if [[ $xpub == *"xpub"* ]]; then
      currency="bitcoin"
    elif [[ $xpub == *"tpub"* ]]; then
      currency="bitcoin_testnet"
    else
      exit 1
    fi

    echo -e ""
    echo -e "${CYAN}SCHEME:${NC}         ${scheme}"
    echo -e "${CYAN}CURRENCY:${NC}       ${currency}"
    echo -e "${CYAN}XPUB:${NC}           ${xpub}"

    cmd="time ledger-live sync --xpub $xpub -c $currency -s $scheme -f stats"

    ledger-live libcoreReset
    echo -e "${CYAN}SYNC NOCACHE:${NC}   explorers.api.live.ledger.com"
    wantnc=$($cmd)
    echo -e "${CYAN}SYNC CACHE:${NC}     explorers.api.live.ledger.com"
    _=$($cmd)  # only useful for benchmarks

    ledger-live libcoreReset
    echo -e "${CYAN}SYNC NOCACHE:${NC}   0.0.0.0:20000"
    gotnc=$(EXPLORER="http://0.0.0.0:20000" $cmd)

    echo -e "${CYAN}SYNC CACHE:${NC}     0.0.0.0:20000"
    gotc=$(EXPLORER="http://0.0.0.0:20000" $cmd)

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
