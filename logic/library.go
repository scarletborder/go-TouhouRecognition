package logic

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	BroadAll             = "全部"
	BroadMainlineDanmaku = "弹幕正作"
	BroadPC98            = "旧作"
)

var broadCategoryWorks = map[string][]string{
	BroadMainlineDanmaku: {
		"东方红魔乡",
		"东方妖妖梦",
		"东方永夜抄",
		"东方花映塚",
		"东方风神录",
		"东方地灵殿",
		"东方星莲船",
		"东方神灵庙",
		"东方辉针城",
		"东方绀珠传",
		"东方天空璋",
		"东方鬼形兽",
		"东方虹龙洞",
		"东方兽王园",
	},
	BroadPC98: {
		"东方灵异传",
		"东方封魔录",
		"东方梦时空",
		"东方幻想乡",
		"东方怪绮谈",
	},
}

type Library struct {
	songs       []Song
	works       []string
	workSet     map[string]struct{}
	songsByWork map[string][]Song
}

func LoadTHWikiLibrary(sourcePath string) (Library, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return Library{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return Library{}, err
	}

	var songs []Song
	workSet := map[string]struct{}{}
	songsByWork := map[string][]Song{}
	for index, record := range records {
		if isCSVHeader(index, record) || isBlankCSVRecord(record) {
			continue
		}
		if len(record) != 3 {
			return Library{}, fmt.Errorf("invalid CSV record at line %d: expected 3 fields, got %d", index+1, len(record))
		}

		song := Song{
			Name:     strings.TrimSpace(record[0]),
			Category: strings.TrimSpace(record[1]),
			URL:      strings.TrimSpace(record[2]),
		}
		if song.Name == "" || song.Category == "" || song.URL == "" {
			return Library{}, fmt.Errorf("empty CSV field at line %d", index+1)
		}
		if !strings.HasPrefix(song.URL, "https://upload.thwiki.cc/") {
			return Library{}, fmt.Errorf("non-THWiki audio URL found for %q: %s", song.Name, song.URL)
		}

		songs = append(songs, song)
		workSet[song.Category] = struct{}{}
		songsByWork[song.Category] = append(songsByWork[song.Category], song)
	}
	if len(songs) == 0 {
		return Library{}, errors.New("no songs parsed from THWiki CSV")
	}

	works := make([]string, 0, len(workSet))
	for work := range workSet {
		works = append(works, work)
	}
	sort.Strings(works)

	return Library{
		songs:       songs,
		works:       works,
		workSet:     workSet,
		songsByWork: songsByWork,
	}, nil
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

func isCSVHeader(index int, record []string) bool {
	return index == 0 &&
		len(record) == 3 &&
		strings.EqualFold(strings.TrimSpace(record[0]), "name") &&
		strings.EqualFold(strings.TrimSpace(record[1]), "category") &&
		strings.EqualFold(strings.TrimSpace(record[2]), "source_url")
}

func isBlankCSVRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}
