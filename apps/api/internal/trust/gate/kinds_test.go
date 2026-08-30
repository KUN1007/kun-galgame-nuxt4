package gate

import "testing"

func TestCanonicalSubjectKindsCoversGateConstants(t *testing.T) {
	set := make(map[string]bool, len(CanonicalSubjectKinds))
	for _, k := range CanonicalSubjectKinds {
		if set[k] {
			t.Fatalf("duplicate kind in CanonicalSubjectKinds: %q", k)
		}
		set[k] = true
	}

	required := []string{
		SubjectKindTopic,
		SubjectKindReply,
		SubjectKindTopicComment,
		SubjectKindTopicPoll,
		SubjectKindTopicLottery,
		SubjectKindGalgameRating,
		SubjectKindGalgameResource,
		SubjectKindGalgameCollection,
		SubjectKindGalgameQuiz,
		SubjectKindGalgameQuizAnswer,
		SubjectKindToolset,
		SubjectKindToolsetResource,
		SubjectKindGalgameComment,
		SubjectKindGalgame,
		SubjectKindUser,
		SubjectKindTodo,
	}
	for _, k := range required {
		if !set[k] {
			t.Errorf("CanonicalSubjectKinds is missing gate constant %q", k)
		}
	}

	if len(CanonicalSubjectKinds) != len(required) {
		t.Errorf("CanonicalSubjectKinds has %d kinds, required set has %d — keep them in lockstep",
			len(CanonicalSubjectKinds), len(required))
	}

	items := CanonicalSubjectKindItems()
	if len(items) != len(CanonicalSubjectKinds) {
		t.Fatalf("adapter length %d != slice length %d", len(items), len(CanonicalSubjectKinds))
	}
	for i, it := range items {
		if it.Key != CanonicalSubjectKinds[i] {
			t.Errorf("adapter item %d = %q, want %q", i, it.Key, CanonicalSubjectKinds[i])
		}
	}
}
