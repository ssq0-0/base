package helpers

import (
	"base/actions"
	"base/actions/types"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func FormatActionSequence(actionSequence []actions.Action, intervals []time.Duration) string {
	var builder strings.Builder

	for i, action := range actionSequence {
		builder.WriteString(fmt.Sprintf("Step %d: %s", i+1, string(action.Type)))

		if action.DexParams != (types.DexParams{}) {
			builder.WriteString(fmt.Sprintf(
				"\n\tFrom Token: %s\n\tTo Token: %s",
				action.DexParams.FromToken.Hex(),
				action.DexParams.ToToken.Hex(),
			))
		}

		if action.RefuelParams != (types.RefuelParams{}) {
			builder.WriteString(fmt.Sprintf(
				"\n\tFrom chain: %s\n\tTo chain: %s",
				action.RefuelParams.ScrChain,
				action.RefuelParams.DstChain,
			))
		}

		if i < len(intervals) {
			builder.WriteString(fmt.Sprintf("\n\tTime to wait: %v", intervals[i]))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

func GetRandomDuration(min, max int) time.Duration {
	if min < 0 || max < 0 {
		min = 20
		max = 40
	}

	if min >= max {
		return time.Duration(min) * time.Minute
	}
	return time.Duration(rand.Intn(max-min+1)+min) * time.Minute
}

func DistributeActionsOverDuration(numActions int, totalDuration time.Duration) []time.Duration {
	if numActions <= 0 {
		return nil
	}

	baseInterval := totalDuration / time.Duration(numActions)

	intervals := make([]time.Duration, numActions)
	for i := 0; i < numActions; i++ {
		variation := float64(baseInterval) * 0.2
		randomVariation := time.Duration(rand.Float64()*2*variation - variation)
		intervals[i] = baseInterval + randomVariation
	}

	return intervals
}
