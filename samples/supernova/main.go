// Supernova operations
package main

import (
	"context"
	"dagger/supernova/internal/dagger"
	"fmt"
	"time"
)

const (
	DEFAULT_CHAINID      = "dev"
	DEFAULT_SUBACCOUNTS  = 1
	DEFAULT_TRANSACTIONS = 10
	MNEMONIC             = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
)

type Supernova struct{}

// Runs a simple Supernova task generating transactions
func (s *Supernova) RunTest(
	ctx context.Context,
	// +optional
	subAccounts int,
	// +optional
	transactions int,
) (int, error) {

	if subAccounts == 0 {
		subAccounts = DEFAULT_SUBACCOUNTS
	}
	if transactions == 0 {
		transactions = DEFAULT_TRANSACTIONS
	}

	gnoSvc := dag.Gnoland().RunGnolandValidator("")

	waitSvc, err := dag.Container().
		From("alpine:3").
		// invalidate cache
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithServiceBinding("gno", gnoSvc).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"curl", "-fsS", "--retry", "5", "--retry-delay", "10", "--retry-all-errors", "http://gno:26657/status?height_gte=1"}).
		ExitCode(ctx)

	if err != nil || waitSvc != 0 {
		return -1, err
	}

	return dag.Container().
		From("ghcr.io/gnolang/supernova:latest").
		// invalidate cache
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithServiceBinding("gno", gnoSvc).
		WithExec([]string{"-sub-accounts", fmt.Sprintf("%d", subAccounts), "-transactions", fmt.Sprintf("%d", transactions),
			"-chain-id", DEFAULT_CHAINID, "-url", "http://gno:26657",
			"-mnemonic", MNEMONIC},
			dagger.ContainerWithExecOpts{
				UseEntrypoint: true,
			}).
		ExitCode(ctx)
}
