package logic

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"strings"
)

func cleanList(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func existingWorksOnly(workSet map[string]struct{}, works []string) []string {
	var result []string
	for _, work := range works {
		if _, ok := workSet[work]; ok {
			result = append(result, work)
		}
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func secureInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be positive")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func secureFloat(max float64) (float64, error) {
	if max <= 0 {
		return 0, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	if err != nil {
		return 0, err
	}
	return (float64(n.Int64()) / 1_000_000_000) * max, nil
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func formatSeconds(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
