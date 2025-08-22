package main

import (
	"fmt"
)

const MaxValidatorCount = 10

func IsValidTopology(valCounter, sentryCounter, sentryRatio int) error {
	if valCounter == 2 {
		return fmt.Errorf("invalid valCounter: %d impossible to reach 2/3 majority", valCounter)
	}

	if valCounter <= 0 || valCounter >= MaxValidatorCount {
		return fmt.Errorf("invalid valCounter: %d (must be between 1 and %d)", valCounter, MaxValidatorCount)
	}

	if sentryRatio < 1 {
		return fmt.Errorf("invalid sentryRatio: %d (must be greater than 0)", sentryRatio)
	}

	if sentryCounter == 0 { // no more checks needed
		return nil
	}

	// Ratio validator
	minVals := (sentryCounter-1)*sentryRatio + 1
	maxVals := sentryCounter * sentryRatio
	if valCounter < minVals || valCounter > maxVals {
		return fmt.Errorf(
			"invalid topology: with %d sentries and ratio %d, validators must be between %d and %d",
			sentryCounter, sentryRatio, minVals, maxVals,
		)
	}

	return nil
}
