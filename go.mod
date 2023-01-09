module github.com/ledgerhq/satstack

go 1.17

require (
	github.com/btcsuite/btcd v0.23.1
	github.com/btcsuite/btcd/btcutil v1.1.2
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1
	github.com/gin-gonic/gin v1.8.1
	github.com/magefile/mage v1.14.0
	github.com/mattn/go-runewidth v0.0.13
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.5.0
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.0 // indirect
	github.com/goccy/go-json v0.9.10 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/btcsuite/btcd v0.21.0-beta.0.20201114000516-e9c7a5ac6401 => github.com/onyb/btcd v0.20.1-beta.0.20201116101952-848ee6a30375
