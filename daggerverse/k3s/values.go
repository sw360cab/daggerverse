package main

const (
	AppTypeHelmKey       = "app.type"
	SvcTypeHelmKey       = "svc.type"
	RpcListenHelmKey     = "gnoland.config.rpc.laddr"
	P2pPexHelmKey        = "gnoland.config.p2p.pex"
	P2pPeersHelmKey      = "gnoland.config.p2p.persistent_peers"
	P2pPrivateIdsHelmKey = "gnoland.config.p2p.private_peer_ids"
	P2pSeedHelmKey       = "gnoland.config.p2p.seeds"
	SvcLoadBalancerValue = "LoadBalancer"
	SvcSuffix            = "-svc"
	P2pPort              = "26656"
	GnolandBinary        = "ghcr.io/gnolang/gno/gnoland:master"
)

var (
	RpcHelmValues = map[string]string{
		AppTypeHelmKey:   "rpc",
		SvcTypeHelmKey:   SvcLoadBalancerValue,
		RpcListenHelmKey: "tcp://0.0.0.0:26657",
	}
	SentryHelmValues = map[string]string{
		P2pPexHelmKey: "true",
		// SvcTypeHelmKey: SvcLoadBalancerValue,
	}
	gnoServices = []gnoService{
		{
			name:      "gnoweb",
			deployDir: "gnoweb",
			port:      8888,
			testPath:  "/r/gnoland",
		},
	}
	rpcService = gnoService{
		name:      "gnocore-rpc-01-svc",
		deployDir: "",
		port:      26657,
		testPath:  "",
	}
)
