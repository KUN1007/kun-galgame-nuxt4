package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The label face is where wave 209's two breaking changes land together: the
// record gained localized{}, and aliases[] stopped being bare strings. A
// consumer that decoded the old shapes gets an empty alias list and a Japanese
// header, both silently — nothing errors.
func TestCatalogLabel_RendersTheChineseNameAndDecodesObjectAliases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"13","display_name":"ねこねこソフト","lang":"ja","company_kind":"game_brand",
			"localized":{"zh-Hans":{"value":"猫猫社","kind":"translation","is_machine":true}},
			"aliases":[
			  {"value":"猫猫社","lang":"zh-Hans","kind":"translation","is_machine":true},
			  {"value":"NekoNeko-soft","lang":"en","kind":"spelling_variant"}],
			"intros":[
			  {"lang":"ja","intro":"げんぶん","source":"bangumi"},
			  {"lang":"zh-Hans","intro":"译文","source":"bangumi","is_machine":true}]}`))
	}))
	t.Cleanup(srv.Close)

	rec, found, movedTo, appErr := New(srv.URL, "test-key", "").
		CatalogLabel(context.Background(), "13")
	if appErr != nil || !found || movedTo != 0 {
		t.Fatalf("label: found=%v movedTo=%d err=%v", found, movedTo, appErr)
	}

	name, original := CatalogEntityNames(context.Background(), rec.Localized, rec.DisplayName, "")
	if name != "猫猫社" || original != "ねこねこソフト" {
		t.Errorf("name/original = %q/%q, want 猫猫社/ねこねこソフト", name, original)
	}
	if len(rec.Aliases) != 2 || rec.Aliases[0].Value != "猫猫社" || rec.Aliases[0].Lang != "zh-Hans" {
		t.Fatalf("aliases = %+v, want the object rows with their language", rec.Aliases)
	}
	if !rec.Aliases[0].Machine {
		t.Error("the machine flag on an alias row was dropped")
	}

	zh := rec.Intros[1]
	if zh.Lang != "zh-Hans" || zh.Intro != "译文" || !zh.Machine {
		t.Errorf("zh intro = %+v, want the Chinese row carrying its machine flag", zh)
	}
}
