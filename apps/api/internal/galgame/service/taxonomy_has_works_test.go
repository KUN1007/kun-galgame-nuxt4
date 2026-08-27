package service

import (
	"context"
	"net/url"
	"testing"
)

func TestTagList_AsksForTagsThatHaveWorks(t *testing.T) {
	rec := &worksQueryRecorder{}
	svc := NewTagService(rec.client(t), &GalgameEnricher{}, nil)

	if _, appErr := svc.GetList(context.Background(), url.Values{}, true); appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if got := rec.get("has_works"); got != "true" {
		t.Fatalf("has_works = %q, want \"true\" — v2 types it as a real boolean and 400s on 1; without it the empty vocabulary is back in the list", got)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Fatalf("nsfw = %q, want true — the index is the open population, SFW filters locally", got)
	}
}

func TestOfficialList_AsksForLabelsThatHaveWorks(t *testing.T) {
	rec := &worksQueryRecorder{}
	svc := NewOfficialService(rec.client(t), nil)

	if _, appErr := svc.GetList(context.Background(), url.Values{}); appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if got := rec.get("has_works"); got != "true" {
		t.Fatalf("has_works = %q, want \"true\" — v2 types it as a real boolean and 400s on 1; without it the empty vocabulary is back in the list", got)
	}
	if got := rec.get("nsfw"); got != "true" {
		t.Fatalf("nsfw = %q, want true", got)
	}
}

func TestOfficialList_KindFilterKeepsHasWorks(t *testing.T) {
	rec := &worksQueryRecorder{}
	svc := NewOfficialService(rec.client(t), nil)

	_, appErr := svc.GetList(context.Background(), url.Values{"kind": {"game_brand"}})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if got := rec.get("kind"); got != "game_brand" {
		t.Fatalf("kind = %q, want \"game_brand\"", got)
	}
	if got := rec.get("has_works"); got != "true" {
		t.Fatalf("has_works = %q, want \"true\"", got)
	}
}
