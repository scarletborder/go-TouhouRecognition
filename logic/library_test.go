package logic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTHWikiLibraryFromCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sourceTHB.csv")
	csvData := `"name","category","source_url"
"A Sacred Lot","东方灵异传","https://upload.thwiki.cc/0/0b/A_Sacret_Lot_0.mp3"
"上海红茶馆　～ Chinese Tea","东方红魔乡","https://upload.thwiki.cc/a/a9/th06_06.mp3"
`
	if err := os.WriteFile(path, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write test CSV: %v", err)
	}

	library, err := LoadTHWikiLibrary(path)
	if err != nil {
		t.Fatalf("LoadTHWikiLibrary returned error: %v", err)
	}
	if library.SongCount() != 2 {
		t.Fatalf("SongCount = %d, want 2", library.SongCount())
	}
	if library.WorkCount() != 2 {
		t.Fatalf("WorkCount = %d, want 2", library.WorkCount())
	}
}

func TestLoadTHWikiLibraryRejectsNonTHWikiURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sourceTHB.csv")
	csvData := `"name","category","source_url"
"bad","东方红魔乡","https://example.com/bad.mp3"
`
	if err := os.WriteFile(path, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write test CSV: %v", err)
	}

	if _, err := LoadTHWikiLibrary(path); err == nil {
		t.Fatal("LoadTHWikiLibrary returned nil error for non-THWiki URL")
	}
}

func TestAllowedWorksExcludesDetailedCategoriesAfterSelection(t *testing.T) {
	library := Library{
		workSet: map[string]struct{}{
			"东方红魔乡": {},
			"东方妖妖梦": {},
			"东方永夜抄": {},
		},
	}
	service := NewService(library)

	works, err := service.allowedWorks(
		[]string{BroadMainlineDanmaku},
		[]string{"东方红魔乡", "东方妖妖梦"},
		[]string{"东方妖妖梦"},
	)
	if err != nil {
		t.Fatalf("allowedWorks returned error: %v", err)
	}
	if _, ok := works["东方红魔乡"]; !ok {
		t.Fatal("allowedWorks excluded 东方红魔乡, want it included")
	}
	if _, ok := works["东方妖妖梦"]; ok {
		t.Fatal("allowedWorks included 东方妖妖梦, want it excluded")
	}
	if len(works) != 1 {
		t.Fatalf("allowedWorks returned %d works, want 1", len(works))
	}
}

func TestAllowedWorksRejectsUnknownExceptDetailedCategory(t *testing.T) {
	library := Library{
		workSet: map[string]struct{}{
			"东方红魔乡": {},
		},
	}
	service := NewService(library)

	if _, err := service.allowedWorks(nil, nil, []string{"不存在的作品"}); err == nil {
		t.Fatal("allowedWorks returned nil error for unknown except detailed category")
	}
}
