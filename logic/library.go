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
	BroadAll                   = "全部"
	BroadMainlineDanmaku       = "弹幕正作"
	BroadPC98                  = "旧作"
	BroadDecimalShootingGames  = "小数点射击游戏"
	BroadTwilightFrontierWorks = "黄昏作品"
	BroadTH06To09              = "TH06-09"
	BroadTH10To12              = "TH10-12"
	BroadTH13To15              = "TH13-15"
	BroadTH16To20              = "TH16-19"
	BroadCD                    = "CD"
	BroadBooks                 = "书籍"
	BroadLenEn                 = "连缘"
)

var broadCategoryOrder = []string{
	BroadAll,
	BroadMainlineDanmaku,
	BroadPC98,
	BroadDecimalShootingGames,
	BroadTwilightFrontierWorks,
	BroadTH06To09,
	BroadTH10To12,
	BroadTH13To15,
	BroadTH16To20,
	BroadCD,
	BroadBooks,
	BroadLenEn,
}

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
	BroadDecimalShootingGames: {
		"东方文花帖",
		"DS东方文花帖",
		"妖精大战争",
		"弹幕天邪鬼",
		"秘封噩梦日记",
		"弹幕狂们的黑市",
	},
	BroadTwilightFrontierWorks: {
		"东方萃梦想",
		"东方绯想天",
		"东方非想天则",
		"东方心绮楼",
		"东方深秘录",
		"东方凭依华",
		"完全凭依唱片名录",
		"深秘乐曲集·补",
		"暗黑能乐集心绮楼",
		"东方刚欲异闻",
	},
	BroadTH06To09: {
		"东方红魔乡",
		"东方妖妖梦",
		"东方永夜抄",
		"东方花映塚",
	},
	BroadTH10To12: {
		"东方风神录",
		"东方地灵殿",
		"东方星莲船",
	},
	BroadTH13To15: {
		"东方神灵庙",
		"东方辉针城",
		"东方绀珠传",
	},
	BroadTH16To20: {
		"东方天空璋",
		"东方鬼形兽",
		"东方虹龙洞",
		"东方兽王园",
		"东方锦上京",
	},
	BroadCD: {
		"蓬莱人形",
		"莲台野夜行",
		"梦违科学世纪",
		"卯酉东海道",
		"大空魔术",
		"未知之花 魅知之旅",
		"鸟船遗迹",
		"伊奘诺物质",
		"燕石博物志",
		"旧约酒馆",
		"虹色的北斗七星",
		"七夕坂梦幻能",
		"灵长新益京",
	},
	BroadBooks: {
		"东方文花帖（书籍）",
		"东方求闻史纪",
		"东方三月精E",
		"东方三月精S1",
		"东方儚月抄（漫画）",
		"东方三月精S2",
		"The Grimoire of Marisa",
		"东方三月精S3",
		"东方三月精O1",
		"东方铃奈庵",
	},
	BroadLenEn: {
		"（连缘）连缘无现里",
		"（连缘）连缘蛇从剑",
		"（连缘）连缘灵烈传",
		"（连缘）连缘天影战记",
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
