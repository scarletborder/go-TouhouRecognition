package logic

import (
	"math"
	"strings"
	"unicode"
)

const (
	minPartialAnswerRunes = 4
	partialMatchScore     = 0.92
)

func (s *Service) VerifyAnswer(req VerifyAnswerRequest) (VerifyAnswerResponse, error) {
	userAnswer := strings.TrimSpace(req.UserAnswer)
	if userAnswer == "" {
		return VerifyAnswerResponse{}, badRequest("user_answer is required")
	}

	correctVariants := answerVariants(req)
	if len(correctVariants) == 0 {
		return VerifyAnswerResponse{}, badRequest("correct_name or correct_answer is required")
	}

	normalizedUser := normalizeAnswer(userAnswer)
	if normalizedUser == "" {
		return VerifyAnswerResponse{}, badRequest("user_answer has no comparable characters")
	}

	best := VerifyAnswerResponse{
		Correct:              false,
		Score:                0,
		MatchedBy:            "none",
		NormalizedUserAnswer: normalizedUser,
	}

	for _, variant := range correctVariants {
		normalizedCorrect := normalizeAnswer(variant)
		if normalizedCorrect == "" {
			continue
		}

		score, matchedBy := compareAnswer(normalizedUser, normalizedCorrect)
		if score > best.Score {
			best.Score = roundScore(score)
			best.MatchedBy = matchedBy
			best.NormalizedCorrectAnswer = normalizedCorrect
		}
		if score >= 0.85 {
			best.Correct = true
		}
	}

	return best, nil
}

func answerVariants(req VerifyAnswerRequest) []string {
	var variants []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			variants = append(variants, value)
		}
	}

	add(req.CorrectName)
	add(req.CorrectAnswer)

	// Add all translate names
	for _, name := range req.TranslateNames {
		add(name)
	}

	if req.CorrectName != "" && req.CorrectCategory != "" {
		add(req.CorrectName + "（" + req.CorrectCategory + "）")
	}
	if name := trimAnswerCategory(req.CorrectAnswer); name != "" {
		add(name)
	}
	return dedupeStrings(variants)
}

func compareAnswer(userAnswer, correctAnswer string) (float64, string) {
	if userAnswer == correctAnswer {
		return 1, "exact"
	}

	userLen := runeLen(userAnswer)
	correctLen := runeLen(correctAnswer)
	if userLen >= minPartialAnswerRunes && strings.Contains(correctAnswer, userAnswer) {
		return partialMatchScore, "partial_contains"
	}
	if correctLen >= minPartialAnswerRunes && strings.Contains(userAnswer, correctAnswer) {
		return partialMatchScore, "contains_correct"
	}

	prefix := commonPrefixRunes(userAnswer, correctAnswer)
	if prefix >= minPartialAnswerRunes && prefix >= min(userLen, correctLen)/2 {
		return 0.88, "prefix"
	}

	suffix := commonSuffixRunes(userAnswer, correctAnswer)
	if suffix >= minPartialAnswerRunes && suffix >= min(userLen, correctLen)/2 {
		return 0.86, "suffix"
	}

	score := similarity(userAnswer, correctAnswer)
	return score, "similarity"
}

func normalizeAnswer(value string) string {
	value = strings.ToLower(value)
	value = strings.NewReplacer(
		"　", "",
		"～", "~",
		"—", "-",
		"－", "-",
		"’", "'",
		"“", "\"",
		"”", "\"",
		"？", "?",
		"！", "!",
	).Replace(value)

	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || isCJK(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func trimAnswerCategory(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, sep := range []string{"（", "(", "【", "["} {
		if idx := strings.LastIndex(value, sep); idx > 0 {
			return strings.TrimSpace(value[:idx])
		}
	}
	return ""
}

func similarity(a, b string) float64 {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 || len(br) == 0 {
		return 0
	}

	distance := levenshtein(ar, br)
	longest := max(len(ar), len(br))
	return 1 - float64(distance)/float64(longest)
}

func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func commonPrefixRunes(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	limit := min(len(ar), len(br))
	for i := 0; i < limit; i++ {
		if ar[i] != br[i] {
			return i
		}
	}
	return limit
}

func commonSuffixRunes(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	limit := min(len(ar), len(br))
	for i := 0; i < limit; i++ {
		if ar[len(ar)-1-i] != br[len(br)-1-i] {
			return i
		}
	}
	return limit
}

func runeLen(value string) int {
	return len([]rune(value))
}

func isCJK(r rune) bool {
	return (r >= '\u3400' && r <= '\u9fff') || (r >= '\uf900' && r <= '\ufaff')
}

func roundScore(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, value := range values {
		key := normalizeAnswer(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}
