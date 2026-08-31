package logic

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const defaultCategoryName = "other"

type Library struct {
	songs              []Song
	works              []string
	workSet            map[string]struct{}
	songsByWork        map[string][]Song
	categories         map[string]string   // work -> broad category
	broadCategories    []string            // sorted list of unique broad categories
	broadCategoryWorks map[string][]string // broad category -> works
}

// LoadLibrary loads music data from three CSV files:
// - musicListPath: music_list.csv with columns (music_name, music_url, translate_names)
// - musicInfoPath: music_info.csv with columns (music_name, original_works, asset_url)
// - categoriesPath: categories.csv with columns (original_works, category)
func LoadLibrary(musicListPath, musicInfoPath, categoriesPath string) (Library, error) {
	// Load music_list.csv
	musicList, err := loadMusicList(musicListPath)
	if err != nil {
		return Library{}, fmt.Errorf("failed to load music_list.csv: %w", err)
	}
	musicListSet := make(map[string]struct{})
	musicNameToTranslate := make(map[string][]string)
	for _, record := range musicList {
		musicListSet[record.MusicName] = struct{}{}
		musicNameToTranslate[record.MusicName] = record.TranslateNames
	}

	// Load music_info.csv
	musicInfo, err := loadMusicInfo(musicInfoPath)
	if err != nil {
		return Library{}, fmt.Errorf("failed to load music_info.csv: %w", err)
	}

	// Load categories.csv
	categories, err := loadCategories(categoriesPath)
	if err != nil {
		return Library{}, fmt.Errorf("failed to load categories.csv: %w", err)
	}
	categoryMap := make(map[string]string)
	for _, record := range categories {
		categoryMap[record.OriginalWorks] = record.Category
	}

	// Preprocess: Filter and build Song list
	// 1. Only keep songs that exist in music_list
	// 2. Remove songs that have no asset_url in any work version
	// 3. Each song-work pair must have asset_url
	songs, works, songsByWork := preprocessMusicData(musicInfo, musicListSet, musicNameToTranslate, categoryMap)

	if len(songs) == 0 {
		return Library{}, errors.New("no valid songs after preprocessing")
	}

	workSet := make(map[string]struct{})
	for work := range songsByWork {
		workSet[work] = struct{}{}
	}

	// Build broad categories from the categories data
	broadCatMap := make(map[string]map[string]struct{})
	for work, broadCat := range categoryMap {
		if _, ok := workSet[work]; !ok {
			continue
		}
		if broadCatMap[broadCat] == nil {
			broadCatMap[broadCat] = make(map[string]struct{})
		}
		broadCatMap[broadCat][work] = struct{}{}
	}

	// Add works without explicit category to "other"
	for work := range workSet {
		if _, ok := categoryMap[work]; !ok {
			broadCat := defaultCategoryName
			if broadCatMap[broadCat] == nil {
				broadCatMap[broadCat] = make(map[string]struct{})
			}
			broadCatMap[broadCat][work] = struct{}{}
		}
	}

	// Convert to sorted lists
	var broadCategories []string
	broadCategoryWorks := make(map[string][]string)
	for broadCat, workMap := range broadCatMap {
		broadCategories = append(broadCategories, broadCat)
		var workList []string
		for work := range workMap {
			workList = append(workList, work)
		}
		sort.Strings(workList)
		broadCategoryWorks[broadCat] = workList
	}
	sort.Strings(broadCategories)

	return Library{
		songs:              songs,
		works:              works,
		workSet:            workSet,
		songsByWork:        songsByWork,
		categories:         categoryMap,
		broadCategories:    broadCategories,
		broadCategoryWorks: broadCategoryWorks,
	}, nil
}

func preprocessMusicData(
	musicInfo []MusicInfoRecord,
	musicListSet map[string]struct{},
	musicNameToTranslate map[string][]string,
	categoryMap map[string]string,
) ([]Song, []string, map[string][]Song) {
	// First pass: collect all valid music_name-work-url combinations
	// and filter by music_list
	type musicWorkURL struct {
		musicName string
		work      string
		url       string
	}
	var validCombinations []musicWorkURL

	musicHasValidVersion := make(map[string]bool)
	for _, info := range musicInfo {
		if _, ok := musicListSet[info.MusicName]; !ok {
			continue // Not in music_list, skip
		}
		if info.AssetURL != "" {
			validCombinations = append(validCombinations, musicWorkURL{
				musicName: info.MusicName,
				work:      info.OriginalWorks,
				url:       info.AssetURL,
			})
			musicHasValidVersion[info.MusicName] = true
		}
	}

	// Build songs and group by work
	var songs []Song
	songsByWork := make(map[string][]Song)
	works := make(map[string]struct{})

	// Track which combinations we've added to avoid duplicates
	added := make(map[string]struct{})

	for _, combo := range validCombinations {
		key := combo.musicName + "|" + combo.work + "|" + combo.url
		if _, ok := added[key]; ok {
			continue
		}
		added[key] = struct{}{}

		song := Song{
			Name:           combo.musicName,
			Category:       combo.work,
			URL:            combo.url,
			TranslateNames: musicNameToTranslate[combo.musicName],
		}
		if cat, ok := categoryMap[combo.work]; ok {
			song.BroadCategory = cat
		} else {
			song.BroadCategory = defaultCategoryName
		}

		songs = append(songs, song)
		songsByWork[combo.work] = append(songsByWork[combo.work], song)
		works[combo.work] = struct{}{}
	}

	// Sort works
	workList := make([]string, 0, len(works))
	for work := range works {
		workList = append(workList, work)
	}
	sort.Strings(workList)

	return songs, workList, songsByWork
}

func loadMusicList(path string) ([]MusicRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []MusicRecord
	for idx, record := range records {
		if isCSVHeader(idx, record, 3) || isBlankCSVRecord(record) {
			continue
		}
		if len(record) != 3 {
			return nil, fmt.Errorf("invalid music_list.csv record at line %d: expected 3 fields, got %d", idx+1, len(record))
		}

		musicName := strings.TrimSpace(record[0])
		musicURL := strings.TrimSpace(record[1])
		translateNamesStr := strings.TrimSpace(record[2])

		if musicName == "" {
			return nil, fmt.Errorf("empty music_name at line %d", idx+1)
		}

		var translateNames []string
		if translateNamesStr != "" {
			for _, name := range strings.Split(translateNamesStr, "|") {
				if trimmed := strings.TrimSpace(name); trimmed != "" {
					translateNames = append(translateNames, trimmed)
				}
			}
		}

		result = append(result, MusicRecord{
			MusicName:      musicName,
			MusicURL:       musicURL,
			TranslateNames: translateNames,
		})
	}

	if len(result) == 0 {
		return nil, errors.New("no records found in music_list.csv")
	}

	return result, nil
}

func loadMusicInfo(path string) ([]MusicInfoRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []MusicInfoRecord
	for idx, record := range records {
		if isCSVHeader(idx, record, 3) || isBlankCSVRecord(record) {
			continue
		}
		if len(record) != 3 {
			return nil, fmt.Errorf("invalid music_info.csv record at line %d: expected 3 fields, got %d", idx+1, len(record))
		}

		musicName := strings.TrimSpace(record[0])
		originalWork := strings.TrimSpace(record[1])
		assetURL := strings.TrimSpace(record[2])

		if musicName == "" || originalWork == "" {
			return nil, fmt.Errorf("empty music_name or original_work at line %d", idx+1)
		}

		result = append(result, MusicInfoRecord{
			MusicName:     musicName,
			OriginalWorks: originalWork,
			AssetURL:      assetURL,
		})
	}

	if len(result) == 0 {
		return nil, errors.New("no records found in music_info.csv")
	}

	return result, nil
}

func loadCategories(path string) ([]CategoryRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var result []CategoryRecord
	for idx, record := range records {
		if isCSVHeader(idx, record, 2) || isBlankCSVRecord(record) {
			continue
		}
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid categories.csv record at line %d: expected 2 fields, got %d", idx+1, len(record))
		}

		originalWork := strings.TrimSpace(record[0])
		category := strings.TrimSpace(record[1])

		if originalWork == "" || category == "" {
			return nil, fmt.Errorf("empty original_work or category at line %d", idx+1)
		}

		result = append(result, CategoryRecord{
			OriginalWorks: originalWork,
			Category:      category,
		})
	}

	if len(result) == 0 {
		return nil, errors.New("no records found in categories.csv")
	}

	return result, nil
}

func (l Library) SongCount() int {
	return len(l.songs)
}

func (l Library) WorkCount() int {
	return len(l.works)
}

func (l Library) Works() []string {
	return append([]string(nil), l.works...)
}

func (l Library) Songs() []Song {
	return append([]Song(nil), l.songs...)
}

func (l Library) Categories() []string {
	return append([]string(nil), l.broadCategories...)
}

func (l Library) CategoryWorks(broadCategory string) []string {
	if works, ok := l.broadCategoryWorks[broadCategory]; ok {
		return append([]string(nil), works...)
	}
	return nil
}

func isCSVHeader(index int, record []string, expectedFields int) bool {
	return index == 0 && len(record) >= expectedFields
}

func isBlankCSVRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
