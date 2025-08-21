package main

const (
	RpcListenYaml     = "gnoland.config.rpc.laddr"
	P2pPexYaml        = "gnoland.config.p2p.pex"
	P2pPeersYaml      = "gnoland.config.p2p.persistent_peers"
	P2pPrivateIdsYaml = "gnoland.config.p2p.private_peer_ids"
	P2pSeedYaml       = "gnoland.config.p2p.seeds"
	P2pPort           = "26656"
	GnolandBinary     = "ghcr.io/gnolang/gno/gnoland:master"
)

var (
	RpcHelmValues = map[string]string{
		RpcListenYaml: "tcp://0.0.0.0:26657",
	}
	SentryHelmValues = map[string]string{
		P2pPexYaml: "true",
	}
	gnoServices = []gnoService{
		{
			name:      "gnoweb",
			deployDir: "core/gnoweb",
			port:      8888,
			testPath:  "/",
		},
		{
			name:      "gnofaucet",
			deployDir: "core/gnofaucet",
			port:      5050,
			testPath:  "/health",
		},
		{
			name:      "tx-indexer",
			deployDir: "core/indexer",
			port:      8546,
			testPath:  "/health",
		},
	}
	rpcService = gnoService{
		name:      "gnocore-rpc-01-svc",
		deployDir: "",
		port:      26657,
		testPath:  "",
	}
)
