# go-TouhouRecognition

`go-TouhouRecognition` 是从 [XiaoGeNekidora/TouhouSongRecognitiveTest](https://github.com/XiaoGeNekidora/TouhouSongRecognitiveTest) 中拆分出来的 Go 语言逻辑模块，旨在为未来的机器人项目提供复用支持。

该模块仅保留了 THWiki 数据源支持，专注于：**曲库加载、分类筛选、随机出题、音频片段裁剪以及答案的模糊验证**。

## 安装

```bash
go get github.com/scarletborder/go-TouhouRecognition
```

### 运行环境要求

| 要求 | 用途 |
| :--- | :--- |
| **Go** | 用于构建和导入模块 |
| `ffprobe` | 获取 THWiki 远程音频的时长信息 |
| `ffmpeg` | 对音频进行随机片段裁剪 |
| **网络访问** | 从 `https://upload.thwiki.cc/` 获取音频 |
| `sourceTHB.csv` | THWiki 曲库数据源文件 |

**安装 FFmpeg 工具：**

```bash
# Windows (使用 winget)
winget install Gyan.FFmpeg

# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt-get install ffmpeg
```

---

## 快速开始

推荐使用根包 `THRecogSvc` 作为入口。原有的 `logic` 包依然保留公开，以兼容旧版集成。

```go
package main

import (
	"context"
	"fmt"
	"log"

	threcog "github.com/scarletborder/go-TouhouRecognition"
)

func main() {
	// 初始化服务
	svc, err := threcog.NewTHRecogSvc("sourceTHB.csv")
	if err != nil {
		log.Fatal(err)
	}

	// 生成题目
	question, err := svc.GenerateQuestion(context.Background(), threcog.QuestionRequest{
		FragmentLength:     10,
		BroadCategories:    []string{threcog.BroadMainlineDanmaku},
		DetailedCategories: []string{"东方红魔乡"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("正确答案：", question.CorrectAnswer.Text)
	fmt.Println("音频格式：", question.Audio.ContentType)
	// Audio.DataBase64 是经过 Base64 编码的 MP3 片段
	// 在发送给客户端或机器人平台前，请先进行解码
}
```

---

## 服务 API

### 初始化服务
```go
// 方式 1：直接通过文件路径创建
svc, err := threcog.NewTHRecogSvc("sourceTHB.csv")

// 方式 2：手动加载库后创建
library, err := threcog.LoadTHWikiLibrary("sourceTHB.csv")
svc := threcog.NewTHRecogSvcFromLibrary(library)
```

### 主要方法
| 方法 | 描述 |
| :--- | :--- |
| `Health()` | 返回数据源状态及已加载歌曲数量 |
| `Categories()` | 返回东方系列的大类与详细分类 |
| `GenerateQuestion(ctx, req)` | 随机选择曲目并返回 MP3 片段题目 |
| `VerifyAnswer(req)` | 使用模糊匹配验证用户答案 |
| `SongCount()` / `WorkCount()` | 返回曲库规模 |
| `Library()` | 获取底层的 `logic.Library` 对象 |

---

## 生成题目

```go
question, err := svc.GenerateQuestion(ctx, threcog.QuestionRequest{
	FragmentLength:           10,
	BroadCategories:          []string{threcog.BroadMainlineDanmaku},
	DetailedCategories:       []string{"东方红魔乡", "东方妖妖梦"},
	ExceptDetailedCategories: []string{"东方妖妖梦"}, // 排除特定作品
})
```

---

## 验证答案

```go
result, err := svc.VerifyAnswer(threcog.VerifyAnswerRequest{
	UserAnswer:    "上海红茶馆",
	CorrectAnswer: question.CorrectAnswer.Text,
})
```

**匹配逻辑：**
验证器会自动处理标点符号、空格、大小写和全角符号。匹配算法包含：完全匹配、包含匹配、前缀/后缀匹配以及 Levenshtein 编辑距离相似度计算。得分 `0.85` 或以上即视为正确。

---

## 错误处理

```go
if threcog.IsBadRequest(err) {
	// 请求错误（如分类不存在、答案为空、时长非法等）
}

if threcog.IsUpstreamError(err) {
	// 上游错误（如 THWiki 访问失败、ffprobe/ffmpeg 执行失败等）
}
```

---

## CSV 文件格式

CSV 文件必须包含 `name`, `category`, `source_url` 三列：

```csv
name,category,source_url
上海红茶馆　～ Chinese Tea,东方红魔乡,https://upload.thwiki.cc/a/a9/th06_06.mp3
```
*所有音频 URL 均会校验，必须以 `https://upload.thwiki.cc/` 开头。*

---

## 旧版 `logic` 包

如果你已经在项目中使用了旧版逻辑，可以继续使用：

```go
import "github.com/scarletborder/go-TouhouRecognition/logic"

library, err := logic.LoadTHWikiLibrary("sourceTHB.csv")
service := logic.NewService(library)
```

---

## 测试

```bash
go test ./...
```