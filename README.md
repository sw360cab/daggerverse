# Gno Dagger

A Playground for Dagger CI mostly related to Gnoland

## Modules

* [`base-ai-agent`](daggerverse/base-ai-agent/): basic AI agent example driven by Dagger's LLM integration, backed locally by Docker Model Runner (DMR)
* [`basic`](daggerverse/basic/): introductory hello-world and grep examples from the Dagger quickstart
* [`build_push`](daggerverse/build_push/): builds a Gnoland container image from a local monorepo checkout and optionally pushes it to a registry
* [`dagger-helm`](daggerverse/dagger-helm/): installs the Dagger engine via its Helm chart on a K3s cluster
* [`docs.gno.land`](daggerverse/docs.gno.land/): builds the `docs.gno.land` Docusaurus site and can preview-publish it to Netlify
* [`gnogenesis`](daggerverse/gnogenesis/): builds Gno binaries (gnoland, gnokey, gnocontribs) and drives `gnogenesis` tooling
* [`gnokey`](daggerverse/gnokey/): plays with `gnokey` basic operations — key generation and sending transactions
* [`gnoland`](daggerverse/gnoland/): clones the Gnoland git repo and runs the project's Go tests
* [`k3s`](daggerverse/k3s/): spins a K3s cluster and deploys the official Gnoland Helm chart (validators, sentries, RPC nodes, Gnoweb, Gnofaucet, Indexer)
* [`kind-dagger`](daggerverse/kind-dagger/): spins a local Kind cluster and installs the Dagger Helm chart inside it
* [`supernova`](daggerverse/supernova/): runs Supernova load-test transactions against a Gnoland node service
