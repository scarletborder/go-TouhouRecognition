package logic

import "testing"

func TestVerifyAnswer(t *testing.T) {
	service := NewService(Library{})

	tests := []struct {
		name    string
		req     VerifyAnswerRequest
		correct bool
	}{
		{
			name: "exact answer",
			req: VerifyAnswerRequest{
				UserAnswer:  "U.N.OWEN就是她吗？",
				CorrectName: "U.N.OWEN就是她吗？",
			},
			correct: true,
		},
		{
			name: "missing punctuation",
			req: VerifyAnswerRequest{
				UserAnswer:  "UNOWEN就是她吗",
				CorrectName: "U.N.OWEN就是她吗？",
			},
			correct: true,
		},
		{
			name: "middle fragment",
			req: VerifyAnswerRequest{
				UserAnswer:  "OWEN就是她",
				CorrectName: "U.N.OWEN就是她吗？",
			},
			correct: true,
		},
		{
			name: "answer text with category",
			req: VerifyAnswerRequest{
				UserAnswer:    "上海红茶馆",
				CorrectAnswer: "上海红茶馆　～ Chinese Tea（东方红魔乡）",
			},
			correct: true,
		},
		{
			name: "wrong answer",
			req: VerifyAnswerRequest{
				UserAnswer:  "恋色Magic",
				CorrectName: "U.N.OWEN就是她吗？",
			},
			correct: false,
		},
		{
			name: "match translate name",
			req: VerifyAnswerRequest{
				UserAnswer:  "Japanese Name",
				CorrectName: "Song Name",
				TranslateNames: []string{
					"Japanese Name",
					"English Name",
					"German Name",
				},
			},
			correct: true,
		},
		{
			name: "partial match on translate name",
			req: VerifyAnswerRequest{
				UserAnswer:  "Japanese",
				CorrectName: "Song Name",
				TranslateNames: []string{
					"Japanese Name",
					"English Name",
				},
			},
			correct: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.VerifyAnswer(tt.req)
			if err != nil {
				t.Fatalf("VerifyAnswer returned error: %v", err)
			}
			if resp.Correct != tt.correct {
				t.Fatalf("Correct = %v, want %v; score=%v matched_by=%s", resp.Correct, tt.correct, resp.Score, resp.MatchedBy)
			}
		})
	}
}
