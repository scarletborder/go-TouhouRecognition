package logic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLibraryFromCSV(t *testing.T) {
	dir := t.TempDir()

	// Create music_list.csv
	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Song A,https://thwiki.cc/Song_A,Song A JP|Song A EN
Song B,https://thwiki.cc/Song_B,Song B JP|Song B EN
Song C,https://thwiki.cc/Song_C,Song C JP|Song C EN
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	// Create music_info.csv
	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Song A,Work A,https://upload.thbwiki.cc/Song_A.mp3
Song A,Work B,
Song B,Work A,https://upload.thbwiki.cc/Song_B.mp3
Song C,Work C,
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	// Create categories.csv
	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
Work A,CD
Work B,PC98
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	library, err := LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("LoadLibrary returned error: %v", err)
	}

	// Song C should be removed (no asset_url)
	// Song A with Work B should be removed (no asset_url)
	// Remaining: Song A (Work A), Song B (Work A)
	if library.SongCount() != 2 {
		t.Fatalf("SongCount = %d, want 2", library.SongCount())
	}

	// Works should be: Work A, Work B, Work C
	// But only Work A and Work B have songs, Work C has no songs with asset_url
	works := library.Works()
	if len(works) != 1 {
		t.Fatalf("Works count = %d, want 1 (only Work A)", len(works))
	}
	if works[0] != "Work A" {
		t.Fatalf("First work = %q, want Work A", works[0])
	}
}

func TestLoadLibraryPreprocessesData(t *testing.T) {
	dir := t.TempDir()

	// Create music_list.csv with one song
	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Good Song,https://thwiki.cc/Good_Song,Good Song JP|Good Song EN
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	// Create music_info.csv with multiple entries
	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Good Song,Work1,https://upload.thbwiki.cc/Good_Song.mp3
Good Song,Work2,
Ignored Song,Work3,https://upload.thbwiki.cc/Ignored_Song.mp3
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	// Create categories.csv
	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
Work1,Category1
Work2,Category1
Work3,Category2
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	library, err := LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("LoadLibrary returned error: %v", err)
	}

	// Only Good Song should be loaded (it's in music_list)
	// Ignored Song should be excluded (not in music_list)
	// Good Song with Work2 should be excluded (no asset_url)
	if library.SongCount() != 1 {
		t.Fatalf("SongCount = %d, want 1", library.SongCount())
	}

	songs := library.Songs()
	if songs[0].Name != "Good Song" {
		t.Fatalf("Song name = %q, want Good Song", songs[0].Name)
	}
	if songs[0].Category != "Work1" {
		t.Fatalf("Song category = %q, want Work1", songs[0].Category)
	}
}

func TestLoadLibraryIncludesTranslateNames(t *testing.T) {
	dir := t.TempDir()

	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Test Song,https://thwiki.cc/Test,Japanese|English|German
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Test Song,TestWork,https://upload.thbwiki.cc/Test.mp3
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
TestWork,TestCat
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	library, err := LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("LoadLibrary returned error: %v", err)
	}

	songs := library.Songs()
	if len(songs) != 1 {
		t.Fatalf("Expected 1 song, got %d", len(songs))
	}

	if len(songs[0].TranslateNames) != 3 {
		t.Fatalf("TranslateNames count = %d, want 3", len(songs[0].TranslateNames))
	}

	expectedNames := []string{"Japanese", "English", "German"}
	for i, expected := range expectedNames {
		if songs[0].TranslateNames[i] != expected {
			t.Fatalf("TranslateNames[%d] = %q, want %q", i, songs[0].TranslateNames[i], expected)
		}
	}
}

func TestLoadLibraryAssignsDefaultCategory(t *testing.T) {
	dir := t.TempDir()

	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Test Song,https://thwiki.cc/Test,
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Test Song,UnknownWork,https://upload.thbwiki.cc/Test.mp3
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
SomeWork,SomeCat
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	library, err := LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("LoadLibrary returned error: %v", err)
	}

	categories := library.Categories()
	if !contains(categories, "other") {
		t.Fatal("Expected 'other' category to be present for uncategorized work")
	}
}

func TestAllowedWorksWithBroadCategories(t *testing.T) {
	dir := t.TempDir()

	musicListPath := filepath.Join(dir, "music_list.csv")
	musicListData := `music_name,music_url,translate_names
Song1,https://thwiki.cc/Song1,
Song2,https://thwiki.cc/Song2,
Song3,https://thwiki.cc/Song3,
`
	if err := os.WriteFile(musicListPath, []byte(musicListData), 0o600); err != nil {
		t.Fatalf("write music_list.csv: %v", err)
	}

	musicInfoPath := filepath.Join(dir, "music_info.csv")
	musicInfoData := `music_name,original_works,asset_url
Song1,Work1,https://upload.thbwiki.cc/Song1.mp3
Song2,Work2,https://upload.thbwiki.cc/Song2.mp3
Song3,Work3,https://upload.thbwiki.cc/Song3.mp3
`
	if err := os.WriteFile(musicInfoPath, []byte(musicInfoData), 0o600); err != nil {
		t.Fatalf("write music_info.csv: %v", err)
	}

	categoriesPath := filepath.Join(dir, "categories.csv")
	categoriesData := `original_works,category
Work1,CategoryA
Work2,CategoryA
Work3,CategoryB
`
	if err := os.WriteFile(categoriesPath, []byte(categoriesData), 0o600); err != nil {
		t.Fatalf("write categories.csv: %v", err)
	}

	library, err := LoadLibrary(musicListPath, musicInfoPath, categoriesPath)
	if err != nil {
		t.Fatalf("LoadLibrary returned error: %v", err)
	}

	service := NewService(library)

	works, err := service.allowedWorks([]string{"CategoryA"}, nil, nil)
	if err != nil {
		t.Fatalf("allowedWorks returned error: %v", err)
	}

	if len(works) != 2 {
		t.Fatalf("allowedWorks returned %d works, want 2", len(works))
	}
	if _, ok := works["Work1"]; !ok {
		t.Fatal("Work1 not found in allowed works")
	}
	if _, ok := works["Work2"]; !ok {
		t.Fatal("Work2 not found in allowed works")
	}
}
