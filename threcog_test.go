package threcog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTHRecogSvc(t *testing.T) {
	dir := t.TempDir()

	// Create music_list.csv
	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Song A,https://thwiki.cc/Song_A,Song A JP|Song A EN
Song B,https://thwiki.cc/Song_B,Song B JP|Song B EN
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	// Create music_info.csv
	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Song A,Work A,https://upload.thbwiki.cc/Song_A.mp3
Song B,Work B,https://upload.thbwiki.cc/Song_B.mp3
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	// Create categories.csv
	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
Work A,Category1
Work B,Category2
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	svc, err := NewTHRecogSvc(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("NewTHRecogSvc returned error: %v", err)
	}

	if svc.SongCount() != 2 {
		t.Fatalf("SongCount = %d, want 2", svc.SongCount())
	}
	if svc.Health().SongCount != 2 {
		t.Fatalf("Health SongCount = %d, want 2", svc.Health().SongCount)
	}
	if len(svc.Categories().DetailedCategories) != 2 {
		t.Fatalf("DetailedCategories length = %d, want 2", len(svc.Categories().DetailedCategories))
	}
}

func TestRootVerifyAnswer(t *testing.T) {
	svc := NewTHRecogSvcFromLibrary(Library{})

	resp, err := svc.VerifyAnswer(VerifyAnswerRequest{
		UserAnswer:  "UNOWEN就是她吗",
		CorrectName: "U.N.OWEN就是她吗？",
	})
	if err != nil {
		t.Fatalf("VerifyAnswer returned error: %v", err)
	}
	if !resp.Correct {
		t.Fatalf("Correct = false; score=%v matched_by=%s", resp.Score, resp.MatchedBy)
	}
}
