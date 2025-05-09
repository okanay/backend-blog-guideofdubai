package utils

// CalculateAICost, token kullanımına göre maliyet hesaplar
func CalculateAICost(tokensUsed int) map[string]any {
	// Fiyatlandırma: Input $0.05, Output $0.20 (milyon token başına)
	inputCost := float64(tokensUsed) * 0.05 / 1000000.0
	outputCost := float64(tokensUsed) * 0.20 / 1000000.0
	totalCost := inputCost + outputCost

	return map[string]any{
		"inputTokens":  tokensUsed,
		"outputTokens": tokensUsed, // Yaklaşık bir değer, gerçek durumda farklı olabilir
		"inputCost":    inputCost,
		"outputCost":   outputCost,
		"totalCost":    totalCost,
	}
}
