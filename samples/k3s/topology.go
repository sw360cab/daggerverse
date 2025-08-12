package main

import (
	"fmt"
)

const MaxValidatorCount = 10

func IsValidTopology(valCounter, sentryCounter, sentryRatio int) error {
	if valCounter <= 0 || valCounter >= MaxValidatorCount {
		return fmt.Errorf("invalid valCounter: %d (must be between 1 and %d)", valCounter, MaxValidatorCount)
	}

	if sentryRatio < 1 {
		return fmt.Errorf("invalid sentryRatio: %d (must be greater than 0)", sentryRatio)
	}

	// Ratio validator
	minVals := sentryCounter*(sentryRatio-1) + 1
	maxVals := sentryCounter * sentryRatio
	if valCounter < minVals || valCounter > maxVals {
		return fmt.Errorf(
			"invalid topology: with %d sentries and ratio %d, validators must be between %d and %d",
			sentryCounter, sentryRatio, minVals, maxVals,
		)
	}

	return nil
}
