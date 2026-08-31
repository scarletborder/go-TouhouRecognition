// Package threcog provides the public entrypoint for Touhou music quiz
// recognition services.
package threcog

import (
	"context"

	"github.com/scarletborder/go-TouhouRecognition/logic"
)

type (
	Song                          = logic.Song
	QuestionRequest               = logic.QuestionRequest
	QuestionResponse              = logic.QuestionResponse
	AnswerPayload                 = logic.AnswerPayload
	AudioPayload                  = logic.AudioPayload
	CategoriesResponse            = logic.CategoriesResponse
	HealthResponse                = logic.HealthResponse
	VerifyAnswerRequest           = logic.VerifyAnswerRequest
	VerifyAnswerResponse          = logic.VerifyAnswerResponse
	Library                       = logic.Library
	WorksForBroadCategoryRequest  = logic.WorksForBroadCategoryRequest
	WorksForBroadCategoryResponse = logic.WorksForBroadCategoryResponse
)

// THRecogSvc is the root package service facade.
//
// The original logic package remains public and compatible. This type exists so
// applications can use the module from its root import path for the common
// workflows: loading music data from CSV files, listing categories, generating questions,
// and verifying answers.
type THRecogSvc struct {
	library Library
	service *logic.Service
}

// NewTHRecogSvc loads music data from three CSV files and returns a ready-to-use service.
//
// Parameters:
//   - musicListPath: path to music_list.csv (music_name, music_url, translate_names)
//   - musicInfoPath: path to music_info.csv (music_name, original_works, asset_url)
//   - categoriesPath: path to categories.csv (original_works, category)
func NewTHRecogSvc(musicListPath, musicInfoPath, categoriesPath string) (*THRecogSvc, error) {
	library, err := logic.LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
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

// LoadLibrary loads music library data from three CSV files.
func LoadLibrary(musicListPath, musicInfoPath, categoriesPath string) (Library, error) {
	return logic.LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
}

func (s *THRecogSvc) Health() HealthResponse {
	return s.service.Health()
}

func (s *THRecogSvc) Categories() CategoriesResponse {
	return s.service.Categories()
}

// GetAllCategories returns all available broad categories
func (s *THRecogSvc) GetAllCategories() []string {
	return s.service.GetAllCategories()
}

// GetAllWorks returns all available works (detailed categories)
func (s *THRecogSvc) GetAllWorks() []string {
	return s.service.GetAllWorks()
}

// GetWorksForBroadCategory returns all works belong to specified category
func (s *THRecogSvc) GetWorksForBroadCategory(req WorksForBroadCategoryRequest) WorksForBroadCategoryResponse {
	return s.library.GetWorksForBroadCategory(req)
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
