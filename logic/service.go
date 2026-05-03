package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	defaultFragmentLen = 10
	maxFragmentLen     = 120
	audioContentType   = "audio/mpeg"
)

type Service struct {
	library Library
}

func NewService(library Library) *Service {
	return &Service{library: library}
}

func (s *Service) Health() HealthResponse {
	return HealthResponse{
		OK:        true,
		Source:    "THWiki",
		SongCount: s.library.SongCount(),
	}
}

func (s *Service) Categories() CategoriesResponse {
	groups := map[string][]string{BroadAll: s.library.Works()}
	for _, broad := range broadCategoryOrder {
		if broad == BroadAll {
			continue
		}
		groups[broad] = existingWorksOnly(s.library.workSet, broadCategoryWorks[broad])
	}

	return CategoriesResponse{
		Source:             "THWiki",
		BroadCategories:    append([]string(nil), broadCategoryOrder...),
		DetailedCategories: s.library.Works(),
		Groups:             groups,
		SongCount:          s.library.SongCount(),
	}
}

func (s *Service) GenerateQuestion(ctx context.Context, req QuestionRequest) (QuestionResponse, error) {
	length := req.FragmentLength
	if length == 0 {
		length = defaultFragmentLen
	}
	if length <= 0 || length > maxFragmentLen {
		return QuestionResponse{}, badRequest(fmt.Sprintf("fragment_length must be > 0 and <= %d", maxFragmentLen))
	}

	candidates, err := s.filterSongs(req)
	if err != nil {
		return QuestionResponse{}, err
	}
	if len(candidates) == 0 {
		return QuestionResponse{}, badRequest("no songs match the requested categories")
	}

	idx, err := secureInt(len(candidates))
	if err != nil {
		return QuestionResponse{}, internalError("failed to choose song")
	}
	song := candidates[idx]

	duration, err := probeDuration(ctx, song.URL)
	if err != nil {
		return QuestionResponse{}, upstreamError(fmt.Sprintf("failed to probe THWiki audio: %v", err))
	}

	start := 0.0
	if duration > length {
		start, err = secureFloat(duration - length)
		if err != nil {
			return QuestionResponse{}, internalError("failed to choose audio start")
		}
		start = round2(start)
	}

	audioBytes, err := cutAudio(ctx, song.URL, start, length)
	if err != nil {
		return QuestionResponse{}, upstreamError(fmt.Sprintf("failed to cut THWiki audio: %v", err))
	}

	answerText := fmt.Sprintf("%s（%s）", song.Name, song.Category)
	return QuestionResponse{
		CorrectAnswer: AnswerPayload{
			Text:     answerText,
			Name:     song.Name,
			Category: song.Category,
		},
		Audio: AudioPayload{
			ContentType:     audioContentType,
			DataBase64:      base64.StdEncoding.EncodeToString(audioBytes),
			DurationSeconds: length,
			StartSeconds:    start,
			SourceURL:       song.URL,
		},
	}, nil
}

func (s *Service) filterSongs(req QuestionRequest) ([]Song, error) {
	allowedWorks, err := s.allowedWorks(req.BroadCategories, req.DetailedCategories, req.ExceptDetailedCategories)
	if err != nil {
		return nil, err
	}

	var songs []Song
	for _, song := range s.library.songs {
		if _, ok := allowedWorks[song.Category]; ok {
			songs = append(songs, song)
		}
	}
	return songs, nil
}

func (s *Service) allowedWorks(broadCategories, detailedCategories, exceptDetailedCategories []string) (map[string]struct{}, error) {
	broadCategories = cleanList(broadCategories)
	detailedCategories = cleanList(detailedCategories)
	exceptDetailedCategories = cleanList(exceptDetailedCategories)

	base := map[string]struct{}{}
	if len(broadCategories) == 0 || contains(broadCategories, BroadAll) {
		for work := range s.library.workSet {
			base[work] = struct{}{}
		}
	} else {
		for _, broad := range broadCategories {
			works, ok := broadCategoryWorks[broad]
			if !ok {
				return nil, badRequest(fmt.Sprintf("unknown broad category %q; allowed: %s", broad, strings.Join(broadCategoryOrder, ", ")))
			}
			for _, work := range works {
				if _, exists := s.library.workSet[work]; exists {
					base[work] = struct{}{}
				}
			}
		}
	}

	for _, exceptDetail := range exceptDetailedCategories {
		if _, ok := s.library.workSet[exceptDetail]; !ok {
			return nil, badRequest(fmt.Sprintf("unknown except detailed category %q", exceptDetail))
		}
	}

	final := base
	if len(detailedCategories) > 0 {
		final = map[string]struct{}{}
		for _, detail := range detailedCategories {
			if _, ok := s.library.workSet[detail]; !ok {
				return nil, badRequest(fmt.Sprintf("unknown detailed category %q", detail))
			}
			if _, inBase := base[detail]; inBase {
				final[detail] = struct{}{}
			}
		}
	}

	for _, exceptDetail := range exceptDetailedCategories {
		delete(final, exceptDetail)
	}
	return final, nil
}
