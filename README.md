`go-TouhouRecognition` 是从 [XiaoGeNekidora/TouhouSongRecognitiveTest](https://github.com/XiaoGeNekidora/TouhouSongRecognitiveTest) 中拆出的 Go 逻辑模块，面向未来机器人项目复用。它只保留 THWiki 源，负责曲库加载、分类筛选、随机出题、音频片段裁剪和答案模糊验证。

# go-TouhouRecognition

Touhou music quiz recognition library for Go. It loads a THWiki CSV song library, lists Touhou work categories, generates random audio-clip questions with `ffprobe`/`ffmpeg`, and verifies user answers with tolerant fuzzy matching.

The recommended entrypoint is the root package `THRecogSvc`. The original `logic` package is still public and remains compatible for older integrations.

## Install

```bash
go get github.com/scarletborder/go-TouhouRecognition
```

Runtime requirements:

| Requirement | Why it is needed |
| --- | --- |
| Go | Build and import the module |
| `ffprobe` | Read remote THWiki audio duration |
| `ffmpeg` | Cut a random MP3 fragment |
| Network access | Fetch audio from `https://upload.thwiki.cc/` |
| `sourceTHB.csv` | THWiki song source file |

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

	threcog "github.com/XiaoGeNekidora/go-TouhouRecognition"
)

func main() {
	svc, err := threcog.NewTHRecogSvc("sourceTHB.csv")
	if err != nil {
		log.Fatal(err)
	}

	question, err := svc.GenerateQuestion(context.Background(), threcog.QuestionRequest{
		FragmentLength:     10,
		BroadCategories:    []string{threcog.BroadMainlineDanmaku},
		DetailedCategories: []string{"东方红魔乡"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(question.CorrectAnswer.Text)
	fmt.Println(question.Audio.ContentType)
	fmt.Println(question.Audio.DataBase64)
}
```

`Audio.DataBase64` is an MP3 fragment encoded as base64. Decode it before sending it to a client, bot platform, or storage service that expects raw audio bytes.

## Service API

Create a service from a CSV file:

```go
svc, err := threcog.NewTHRecogSvc("sourceTHB.csv")
```

Or load the library yourself:

```go
library, err := threcog.LoadTHWikiLibrary("sourceTHB.csv")
svc := threcog.NewTHRecogSvcFromLibrary(library)
```

Main methods:

| Method | Description |
| --- | --- |
| `Health()` | Returns source and loaded song count |
| `Categories()` | Returns broad and detailed Touhou work categories |
| `GenerateQuestion(ctx, req)` | Chooses a song and returns an MP3 question fragment |
| `VerifyAnswer(req)` | Checks a user answer against the correct answer |
| `SongCount()` / `WorkCount()` | Returns library size |
| `Songs()` / `Works()` | Returns copied song/work slices |
| `Library()` | Returns the loaded `logic.Library` |

Broad category constants:

```go
threcog.BroadAll
threcog.BroadMainlineDanmaku
threcog.BroadPC98
threcog.BroadDecimalShootingGames
threcog.BroadTwilightFrontierWorks
threcog.BroadTH06To09
threcog.BroadTH10To12
threcog.BroadTH13To15
threcog.BroadTH16To19
threcog.BroadCD
threcog.BroadBooks
threcog.BroadLenEn
```

## Generate a Question

```go
question, err := svc.GenerateQuestion(ctx, threcog.QuestionRequest{
	FragmentLength:           10,
	BroadCategories:          []string{threcog.BroadMainlineDanmaku},
	DetailedCategories:       []string{"东方红魔乡", "东方妖妖梦"},
	ExceptDetailedCategories: []string{"东方妖妖梦"},
})
```

Request fields:

| Field | Description |
| --- | --- |
| `FragmentLength` | Audio length in seconds. Default `10`, max `120` |
| `BroadCategories` | Empty or `BroadAll` means all works |
| `DetailedCategories` | Work names that exist in the CSV, such as `东方红魔乡` |
| `ExceptDetailedCategories` | Work names to remove from the final selected works |

Useful response fields:

```go
question.CorrectAnswer.Text
question.CorrectAnswer.Name
question.CorrectAnswer.Category
question.Audio.ContentType
question.Audio.DataBase64
question.Audio.DurationSeconds
question.Audio.StartSeconds
question.Audio.SourceURL
```

## Verify an Answer

```go
result, err := svc.VerifyAnswer(threcog.VerifyAnswerRequest{
	UserAnswer:  "OWEN就是她",
	CorrectName: question.CorrectAnswer.Name,
})
```

You can also verify against the full answer text:

```go
result, err := svc.VerifyAnswer(threcog.VerifyAnswerRequest{
	UserAnswer:    "上海红茶馆",
	CorrectAnswer: question.CorrectAnswer.Text,
})
```

`CorrectName` or `CorrectAnswer` must be provided. The verifier normalizes punctuation, spaces, case, and common full-width symbols, then tries exact, contains, prefix/suffix, and Levenshtein similarity matching. A score of `0.85` or higher is treated as correct.

## Error Handling

```go
if threcog.IsBadRequest(err) {
	// Invalid category, empty answer, invalid fragment length, and so on.
}

if threcog.IsUpstreamError(err) {
	// THWiki, ffprobe, ffmpeg, or network failure.
}
```

## CSV Format

The CSV file must contain `name`, `category`, and `source_url` columns:

```csv
name,category,source_url
上海红茶馆　～ Chinese Tea,东方红魔乡,https://upload.thwiki.cc/a/a9/th06_06.mp3
```

All audio URLs are validated and must start with `https://upload.thwiki.cc/`.

## Legacy `logic` Package

Existing code can keep using the original subpackage:

```go
import "github.com/XiaoGeNekidora/go-TouhouRecognition/logic"

library, err := logic.LoadTHWikiLibrary("sourceTHB.csv")
service := logic.NewService(library)
```

No existing public `logic` functions or types were removed.

## Test

```bash
go test ./...
```
