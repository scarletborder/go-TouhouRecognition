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
