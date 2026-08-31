[简体中文](./README_CN.md)

# go-TouhouRecognition

Touhou music quiz recognition library for Go. It loads music data from three CSV files (music list, music info, and categories), lists work categories, generates random audio-clip questions with `ffprobe`/`ffmpeg`, and verifies user answers with tolerant fuzzy matching.

The recommended entrypoint is the root package `THRecogSvc`. The original `logic` package is still public and remains compatible for older integrations.

## Install

```bash
go get github.com/scarletborder/go-TouhouRecognition
```

Runtime requirements:

| Requirement | Why it is needed |
| --- | --- |
| Go | Build and import the module |
| `ffprobe` | Read remote audio duration |
| `ffmpeg` | Cut a random MP3 fragment |
| Network access | Fetch audio from `https://upload.thbwiki.cc/` |
| `music_list.csv` | List of music names and translations |
| `music_info.csv` | Music versions and asset URLs |
| `categories.csv` | Work to category mapping |

Install FFmpeg tools:

```bash
# Windows, with winget
winget install Gyan.FFmpeg

# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get install ffmpeg
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	threcog "github.com/scarletborder/go-TouhouRecognition"
)

func main() {
	svc, err := threcog.NewTHRecogSvc(
		"data/music_list.csv",
		"data/music_info.csv",
		"data/categories.csv",
	)
	if err != nil {
		log.Fatal(err)
	}

	// List all available categories
	allCategories := svc.GetAllCategories()
	fmt.Println("Available categories:", allCategories)

	question, err := svc.GenerateQuestion(context.Background(), threcog.QuestionRequest{
		FragmentLength:  10,
		BroadCategories: []string{}, // empty = all categories
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(question.CorrectAnswer.Text)
	fmt.Println(question.CorrectAnswer.TranslateNames)
	fmt.Println(question.Audio.ContentType)
	fmt.Println(question.Audio.DataBase64)
}
```

`Audio.DataBase64` is an MP3 fragment encoded as base64. Decode it before sending it to a client, bot platform, or storage service that expects raw audio bytes.

## Service API

Create a service from three CSV files:

```go
svc, err := threcog.NewTHRecogSvc(
	"data/music_list.csv",
	"data/music_info.csv",
	"data/categories.csv",
)
```

Or load the library yourself:

```go
library, err := threcog.LoadLibrary(
	"data/music_list.csv",
	"data/music_info.csv",
	"data/categories.csv",
)
svc := threcog.NewTHRecogSvcFromLibrary(library)
```

Main methods:

| Method | Description |
| --- | --- |
| `Health()` | Returns source and loaded song count |
| `Categories()` | Returns broad and detailed work categories with groups |
| `GetAllCategories()` | Returns all available broad categories |
| `GetAllWorks()` | Returns all available works (detailed categories) |
| `GenerateQuestion(ctx, req)` | Chooses a song and returns an MP3 question fragment |
| `VerifyAnswer(req)` | Checks a user answer against the correct answer |
| `SongCount()` / `WorkCount()` | Returns library size |
| `Songs()` / `Works()` | Returns copied song/work slices |
| `Library()` | Returns the loaded `logic.Library` |

## Generate a Question

```go
// Get available categories
categories := svc.GetAllCategories()
works := svc.GetAllWorks()

question, err := svc.GenerateQuestion(ctx, threcog.QuestionRequest{
	FragmentLength:           10,
	BroadCategories:          []string{"CD", "PC98"},  // From GetAllCategories()
	DetailedCategories:       []string{"东方红魔乡", "东方妖妖梦"},
	ExceptDetailedCategories: []string{"东方妖妖梦"},
})
```

Request fields:

| Field | Description |
| --- | --- |
| `FragmentLength` | Audio length in seconds. Default `10`, max `120` |
| `BroadCategories` | Empty or omitted means all categories. Values from `GetAllCategories()` |
| `DetailedCategories` | Work names from `GetAllWorks()`, such as `东方红魔乡` |
| `ExceptDetailedCategories` | Work names to exclude from selection |

Useful response fields:

```go
question.CorrectAnswer.Text           // "Song Name (Work Name)"
question.CorrectAnswer.Name           // Song name from music_list
question.CorrectAnswer.Category       // Work name from music_info
question.CorrectAnswer.TranslateNames // Alternative names from music_list
question.Audio.ContentType
question.Audio.DataBase64
question.Audio.DurationSeconds
question.Audio.StartSeconds
question.Audio.SourceURL
```

## Verify an Answer

```go
result, err := svc.VerifyAnswer(threcog.VerifyAnswerRequest{
	UserAnswer:     "OWEN就是她",
	CorrectName:    question.CorrectAnswer.Name,
	TranslateNames: question.CorrectAnswer.TranslateNames,
})
```

You can also verify against the full answer text:

```go
result, err := svc.VerifyAnswer(threcog.VerifyAnswerRequest{
	UserAnswer:     "上海红茶馆",
	CorrectAnswer:  question.CorrectAnswer.Text,
	TranslateNames: question.CorrectAnswer.TranslateNames,
})
```

The verifier tries to match against:
- The exact `CorrectName` (e.g., "U.N.OWEN就是她吗？")
- All `TranslateNames` provided (e.g., "U.N.Owen Was It Just A Dream～", "Just a Dream～", "うんおーえん")
- The full `CorrectAnswer` text with category

It normalizes punctuation, spaces, case, and common full-width symbols, then tries exact, contains, prefix/suffix, and Levenshtein similarity matching. A score of `0.85` or higher is treated as correct.

## Error Handling

```go
if threcog.IsBadRequest(err) {
	// Invalid category, empty answer, invalid fragment length, and so on.
}

if threcog.IsUpstreamError(err) {
	// Audio fetch, ffprobe, ffmpeg, or network failure.
}
```

## CSV Files Format

### music_list.csv

Contains music names and their alternative translations:

```csv
music_name,music_url,translate_names
上海红茶馆　～ Chinese Tea,https://thwiki.cc/...,Shanghai Alice Teahouse|红茶館 〜 Chinese Tea|Chinese Tea
U.N.OWEN就是她吗？,https://thwiki.cc/...,U.N.Owen Was It Just A Dream～|Just a Dream～|うんおーえん
```

- **music_name**: Unique identifier for each song (required)
- **music_url**: URL to song page on THWiki (for reference, optional)
- **translate_names**: Pipe-separated alternative names for the song (used in answer verification)

### music_info.csv

Maps each song to its versions across different works with asset URLs:

```csv
music_name,original_works,asset_url
上海红茶馆　～ Chinese Tea,东方红魔乡,https://upload.thbwiki.cc/a/a9/th06_06.mp3
U.N.OWEN就是她吗？,东方灵异传,https://upload.thbwiki.cc/0/0b/A_Sacret_Lot_0.mp3
U.N.OWEN就是她吗？,东方妖妖梦,https://upload.thbwiki.cc/.../th7_06.mp3
```

- **music_name**: Must exist in music_list.csv
- **original_works**: Work/game name for this version
- **asset_url**: Playable MP3 URL (must be empty for versions that don't have a recorded version)

**Important**: 
- A song with all versions having empty `asset_url` will be skipped entirely
- A song version with empty `asset_url` will be ignored during loading

### categories.csv

Maps works to their broad categories:

```csv
original_works,category
东方红魔乡,Mainline Danmaku
东方妖妖梦,Mainline Danmaku
东方灵异传,PC98
东方幻想乡,PC98
```

- **original_works**: Must match `original_works` from music_info.csv
- **category**: Broad category name (users can group/filter by this)
- Works not listed here are automatically assigned to category `"other"`

## Data Loading Flow

1. Load `music_list.csv` → get all music names and translate names
2. Load `music_info.csv` → filter by music_list, remove entries with no asset_url
3. Load `categories.csv` → map works to broad categories
4. Build Library → organize songs by work, works by category
5. Ready for question generation and answer verification

## Test

```bash
go test ./...
```
