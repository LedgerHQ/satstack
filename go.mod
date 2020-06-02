module ledger-sats-stack

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta.0.20200414114020-8b54b0b96418
	github.com/btcsuite/btcutil v1.0.2
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.1.1
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	go.etcd.io/bbolt v1.3.4
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	golang.org/x/tools v0.0.0-20200530233709-52effbd89c51 // indirect
)

// Uncomment the line below to enable experimental JSON-RPC features.
replace github.com/btcsuite/btcd v0.20.1-beta.0.20200414114020-8b54b0b96418 => github.com/onyb/btcd v0.20.1-beta.0.20200525154529-e30a38b3b08d
