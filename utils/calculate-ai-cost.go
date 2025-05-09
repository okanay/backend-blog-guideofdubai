package utils

import "fmt"

func CalculateAICost(tokensUsed int) map[string]any {
	// Pricing: Input $0.20, Output $0.80 (per million tokens) 4.1 mini
	// Pricing: Input $0.05, Output $0.20 (per million tokens) 4.1 nano

	inputCost := float64(tokensUsed) * 0.05 / 1000000.0
	outputCost := float64(tokensUsed) * 0.20 / 1000000.0
	totalCost := inputCost + outputCost

	return map[string]any{
		"inputTokens":  tokensUsed,
		"outputTokens": tokensUsed,
		"inputCost":    fmt.Sprintf("$%.4f", inputCost),
		"outputCost":   fmt.Sprintf("$%.4f", outputCost),
		"totalCost":    fmt.Sprintf("$%.4f", totalCost),
	}
}
