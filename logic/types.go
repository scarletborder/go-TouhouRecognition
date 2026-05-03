package logic

type Song struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	URL      string `json:"source_url"`
}

type QuestionRequest struct {
	FragmentLength           float64  `json:"fragment_length"`
	BroadCategories          []string `json:"broad_categories"`
	DetailedCategories       []string `json:"detailed_categories"`
	ExceptDetailedCategories []string `json:"except_detailed_categories"`
}

type QuestionResponse struct {
	CorrectAnswer AnswerPayload `json:"correct_answer"`
	Audio         AudioPayload  `json:"audio"`
}

type AnswerPayload struct {
	Text     string `json:"text"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type AudioPayload struct {
	ContentType     string  `json:"content_type"`
	DataBase64      string  `json:"data_base64"`
	DurationSeconds float64 `json:"duration_seconds"`
	StartSeconds    float64 `json:"start_seconds"`
	SourceURL       string  `json:"source_url"`
}

type CategoriesResponse struct {
	Source             string              `json:"source"`
	BroadCategories    []string            `json:"broad_categories"`
	DetailedCategories []string            `json:"detailed_categories"`
	Groups             map[string][]string `json:"groups"`
	SongCount          int                 `json:"song_count"`
}

type HealthResponse struct {
	OK        bool   `json:"ok"`
	Source    string `json:"source"`
	SongCount int    `json:"song_count"`
}

type VerifyAnswerRequest struct {
	UserAnswer      string `json:"user_answer"`
	CorrectAnswer   string `json:"correct_answer"`
	CorrectName     string `json:"correct_name"`
	CorrectCategory string `json:"correct_category"`
}

type VerifyAnswerResponse struct {
	Correct                 bool    `json:"correct"`
	Score                   float64 `json:"score"`
	MatchedBy               string  `json:"matched_by"`
	NormalizedUserAnswer    string  `json:"normalized_user_answer"`
	NormalizedCorrectAnswer string  `json:"normalized_correct_answer"`
}
