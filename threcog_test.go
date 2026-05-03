package threcog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTHRecogSvc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sourceTHB.csv")
	csvData := `"name","category","source_url"
"A Sacred Lot","东方灵异传","https://upload.thwiki.cc/0/0b/A_Sacret_Lot_0.mp3"
"上海红茶馆　～ Chinese Tea","东方红魔乡","https://upload.thwiki.cc/a/a9/th06_06.mp3"
`
	if err := os.WriteFile(path, []byte(csvData), 0o600); err != nil {
		t.Fatalf("write test CSV: %v", err)
	}

	svc, err := NewTHRecogSvc(path)
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
