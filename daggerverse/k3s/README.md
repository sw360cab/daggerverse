# K3S Cluster Test

Generates a Kubernetes cluster in K3s then resources are deployed and scheduled on it.
Leveraging a Helm Template, it spins network nodes (validators RPC nodes) and collateral services (Gnoweb).

- Simple launch

```sh
dagger call -i -v spin-cluster --data-folder ./data
```
