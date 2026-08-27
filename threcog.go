// Package threcog provides the public entrypoint for Touhou music quiz
// recognition services.
package threcog

import (
	"context"

	"github.com/XiaoGeNekidora/go-TouhouRecognition/logic"
)

const (
	BroadAll                   = logic.BroadAll
	BroadMainlineDanmaku       = logic.BroadMainlineDanmaku
	BroadPC98                  = logic.BroadPC98
	BroadDecimalShootingGames  = logic.BroadDecimalShootingGames
	BroadTwilightFrontierWorks = logic.BroadTwilightFrontierWorks
	BroadTH06To09              = logic.BroadTH06To09
	BroadTH10To12              = logic.BroadTH10To12
	BroadTH13To15              = logic.BroadTH13To15
	BroadTH16To20              = logic.BroadTH16To20
	BroadCD                    = logic.BroadCD
	BroadBooks                 = logic.BroadBooks
	BroadLenEn                 = logic.BroadLenEn
)

type (
	Song                 = logic.Song
	QuestionRequest      = logic.QuestionRequest
	QuestionResponse     = logic.QuestionResponse
	AnswerPayload        = logic.AnswerPayload
	AudioPayload         = logic.AudioPayload
	CategoriesResponse   = logic.CategoriesResponse
	HealthResponse       = logic.HealthResponse
	VerifyAnswerRequest  = logic.VerifyAnswerRequest
	VerifyAnswerResponse = logic.VerifyAnswerResponse
	Library              = logic.Library
)

// THRecogSvc is the root package service facade.
//
// The original logic package remains public and compatible. This type exists so
// applications can use the module from its root import path for the common
// workflows: loading the THWiki CSV, listing categories, generating questions,
// and verifying answers.
type THRecogSvc struct {
	library Library
	service *logic.Service
}

// NewTHRecogSvc loads a THWiki CSV library from sourcePath and returns a ready
// to use recognition service.
func NewTHRecogSvc(sourcePath string) (*THRecogSvc, error) {
	library, err := logic.LoadTHWikiLibrary(sourcePath)
	if err != nil {
		return nil, err
	}
	return NewTHRecogSvcFromLibrary(library), nil
}

// NewTHRecogSvcFromLibrary creates a service from a preloaded library.
func NewTHRecogSvcFromLibrary(library Library) *THRecogSvc {
	return &THRecogSvc{
		library: library,
		service: logic.NewService(library),
	}
}

// LoadTHWikiLibrary keeps the common loader available from the root package.
func LoadTHWikiLibrary(sourcePath string) (Library, error) {
	return logic.LoadTHWikiLibrary(sourcePath)
}

func (s *THRecogSvc) Health() HealthResponse {
	return s.service.Health()
}

func (s *THRecogSvc) Categories() CategoriesResponse {
	return s.service.Categories()
}

func (s *THRecogSvc) GenerateQuestion(ctx context.Context, req QuestionRequest) (QuestionResponse, error) {
	return s.service.GenerateQuestion(ctx, req)
}

func (s *THRecogSvc) VerifyAnswer(req VerifyAnswerRequest) (VerifyAnswerResponse, error) {
	return s.service.VerifyAnswer(req)
}

func (s *THRecogSvc) SongCount() int {
	return s.library.SongCount()
}

func (s *THRecogSvc) WorkCount() int {
	return s.library.WorkCount()
}

func (s *THRecogSvc) Works() []string {
	return s.library.Works()
}

func (s *THRecogSvc) Songs() []Song {
	return s.library.Songs()
}

func (s *THRecogSvc) Library() Library {
	return s.library
}

func IsBadRequest(err error) bool {
	return logic.IsBadRequest(err)
}

func IsUpstreamError(err error) bool {
	return logic.IsUpstreamError(err)
}
