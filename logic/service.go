package logic

import (
	"context"
	"encoding/base64"
	"fmt"
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
		Source:    "Dynamic CSV Data",
		SongCount: s.library.SongCount(),
	}
}

// GetAllCategories returns all available broad categories
func (s *Service) GetAllCategories() []string {
	return s.library.Categories()
}

// GetAllWorks returns all available works (detailed categories)
func (s *Service) GetAllWorks() []string {
	return s.library.Works()
}

func (s *Service) Categories() CategoriesResponse {
	allWorks := s.library.Works()
	groups := make(map[string][]string)

	// Add "all" group with all works
	groups["all"] = append([]string(nil), allWorks...)

	// Add category-based groups
	for _, broadCat := range s.library.Categories() {
		works := s.library.CategoryWorks(broadCat)
		if len(works) > 0 {
			groups[broadCat] = works
		}
	}

	return CategoriesResponse{
		Source:             "Dynamic CSV Data",
		BroadCategories:    s.library.Categories(),
		DetailedCategories: allWorks,
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
		return QuestionResponse{}, upstreamError(fmt.Sprintf("failed to probe THBWiki audio: %v", err))
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
		return QuestionResponse{}, upstreamError(fmt.Sprintf("failed to cut THBWiki audio: %v", err))
	}

	answerText := fmt.Sprintf("%s（%s）", song.Name, song.Category)
	return QuestionResponse{
		CorrectAnswer: AnswerPayload{
			Text:           answerText,
			Name:           song.Name,
			Category:       song.Category,
			TranslateNames: song.TranslateNames,
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
	if len(broadCategories) == 0 || contains(broadCategories, "all") {
		// Include all available works
		for work := range s.library.workSet {
			base[work] = struct{}{}
		}
	} else {
		// Include works from specified broad categories
		for _, broad := range broadCategories {
			works := s.library.CategoryWorks(broad)
			if len(works) == 0 {
				// Check if it's a valid category
				validCats := s.library.Categories()
				if !contains(validCats, broad) {
					return nil, badRequest(fmt.Sprintf("unknown broad category %q", broad))
				}
				continue
			}
			for _, work := range works {
				base[work] = struct{}{}
			}
		}
	}

	// Check that exception categories exist
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

	// Remove exceptions
	for _, exceptDetail := range exceptDetailedCategories {
		delete(final, exceptDetail)
	}

	return final, nil
}
