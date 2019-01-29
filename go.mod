module github.com/endurio/ndraddrgen

require (
	github.com/endurio/ndrd/chaincfg v1.2.2-0.20190127104041-05203c32f47c
	github.com/endurio/ndrd/dcrec v0.0.0-20181212181811-1a370d38d671
	github.com/endurio/ndrd/dcrec/secp256k1 v1.0.1
	github.com/endurio/ndrd/dcrutil v1.2.0
	github.com/endurio/ndrd/hdkeychain v1.1.1
	github.com/endurio/ndrw/walletseed v1.0.1
	golang.org/x/crypto v0.0.0-20190128193316-c7b33c32a30b // indirect
)

replace (
	github.com/endurio => ../
	github.com/endurio/ndraddrgen => ./
	github.com/endurio/ndrd => ../ndrd
	github.com/endurio/ndrd/addrmgr => ../ndrd/addrmgr
	github.com/endurio/ndrd/blockchain => ../ndrd/blockchain
	github.com/endurio/ndrd/certgen => ../ndrd/certgen
	github.com/endurio/ndrd/chaincfg => ../ndrd/chaincfg
	github.com/endurio/ndrd/chaincfg/chainhash => ../ndrd/chaincfg/chainhash
	github.com/endurio/ndrd/connmgr => ../ndrd/connmgr
	github.com/endurio/ndrd/database => ../ndrd/database
	github.com/endurio/ndrd/dcrec => ../ndrd/dcrec
	github.com/endurio/ndrd/dcrec/edwards => ../ndrd/dcrec/edwards
	github.com/endurio/ndrd/dcrec/secp256k1 => ../ndrd/dcrec/secp256k1
	github.com/endurio/ndrd/dcrjson => ../ndrd/dcrjson
	github.com/endurio/ndrd/dcrutil => ../ndrd/dcrutil
	github.com/endurio/ndrd/fees => ../ndrd/fees
	github.com/endurio/ndrd/gcs => ../ndrd/gcs
	github.com/endurio/ndrd/hdkeychain => ../ndrd/hdkeychain
	github.com/endurio/ndrd/limits => ../ndrd/limits
	github.com/endurio/ndrd/mempool => ../ndrd/mempool
	github.com/endurio/ndrd/mining => ../ndrd/mining
	github.com/endurio/ndrd/peer => ../ndrd/peer
	github.com/endurio/ndrd/rpcclient => ../ndrd/rpcclient
	github.com/endurio/ndrd/txscript => ../ndrd/txscript
	github.com/endurio/ndrd/wire => ../ndrd/wire
	github.com/endurio/ndrw => ../ndrw
	github.com/endurio/ndrw/chain => ../ndrw/chain
	github.com/endurio/ndrw/deployments => ../ndrw/deployments
	github.com/endurio/ndrw/errors => ../ndrw/errors
	github.com/endurio/ndrw/internal/helpers => ../ndrw/internal/helpers
	github.com/endurio/ndrw/internal/zero => ../ndrw/internal/zero
	github.com/endurio/ndrw/lru => ../ndrw/lru
	github.com/endurio/ndrw/p2p => ../ndrw/p2p
	github.com/endurio/ndrw/pgpwordlist => ../ndrw/pgpwordlist
	github.com/endurio/ndrw/rpc/walletrpc => ../ndrw/rpc/walletrpc
	github.com/endurio/ndrw/spv => ../ndrw/spv
	github.com/endurio/ndrw/validate => ../ndrw/validate
	github.com/endurio/ndrw/version => ../ndrw/version
	github.com/endurio/ndrw/wallet => ../ndrw/wallet
	github.com/endurio/ndrw/walletseed => ../ndrw/walletseed
)
