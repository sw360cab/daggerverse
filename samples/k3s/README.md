# K3S Cluster Test

Generates a Kubernetes cluster in K3s and deploy and schedule resources in it, leveraging HELM and official Gnoland Infra repository it spins a network of nodes (validators, sentries and RPC nodes) and collateral services (Gnoweb, Gnofaucet, Indexer).

- Simple launch

```sh
dagger call -i -v spin-cluster --helm-data-folder helm-values --repo-token "op://<Path_to_token_in_1Password>"
```

- Complex network

```sh
dagger call -i -v spin-cluster --helm-data-folder helm-values --repo-token "op://<Path_to_token_in_1Password>" --val-counter 6 --sentry-counter 2 --sentry-ratio 3
```
