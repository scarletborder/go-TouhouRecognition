package logic

// Song represents a music piece with its work and category information
type Song struct {
	Name           string   `json:"name"`            // music_name
	Category       string   `json:"category"`        // original_work
	URL            string   `json:"source_url"`      // asset_url
	TranslateNames []string `json:"translate_names"` // alternative names for the song
	BroadCategory  string   `json:"broad_category"`  // broad category from categories.csv
}

// MusicRecord represents a row from music_list.csv
type MusicRecord struct {
	MusicName      string
	MusicURL       string
	TranslateNames []string // pipe-separated, will be split
}

// MusicInfoRecord represents a row from music_info.csv
type MusicInfoRecord struct {
	MusicName     string
	OriginalWorks string
	AssetURL      string
}

// CategoryRecord represents a row from categories.csv
type CategoryRecord struct {
	OriginalWorks string
	Category      string
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
	Text           string   `json:"text"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	TranslateNames []string `json:"translate_names"`
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
	UserAnswer      string   `json:"user_answer"`
	CorrectAnswer   string   `json:"correct_answer"`
	CorrectName     string   `json:"correct_name"`
	CorrectCategory string   `json:"correct_category"`
	TranslateNames  []string `json:"translate_names"` // alternative names for the music
}

type VerifyAnswerResponse struct {
	Correct                 bool    `json:"correct"`
	Score                   float64 `json:"score"`
	MatchedBy               string  `json:"matched_by"`
	NormalizedUserAnswer    string  `json:"normalized_user_answer"`
	NormalizedCorrectAnswer string  `json:"normalized_correct_answer"`
}
