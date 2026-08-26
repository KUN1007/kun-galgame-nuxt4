package client

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRewriteV2JSON_AttributionRoleBecomesRole(t *testing.T) {
	raw := []byte(`{"object":"company","id":"13","display_name":"FAVORITE","company_kind":"game_brand","attribution_role":"developer"}`)
	var got map[string]any
	if err := json.Unmarshal(rewriteV2JSON(raw, ""), &got); err != nil {
		t.Fatalf("rewrite produced invalid json: %v", err)
	}
	if got["role"] != "developer" {
		t.Fatalf("role = %v, want the attribution role, not the company kind", got["role"])
	}
	if got["kind"] != "game_brand" || got["label_kind"] != "game_brand" {
		t.Fatalf("company_kind mapping regressed: %v", got)
	}
}

func TestMakerName_PrefersTheRoleThatMadeTheGame(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name   string
		labels []catWorkLabel
		want   string
	}{
		{
			name: "developer wins over a publisher listed first",
			labels: []catWorkLabel{
				{DisplayName: "Dramatic Create", Role: "publisher"},
				{DisplayName: "FAVORITE", Role: "developer"},
			},
			want: "FAVORITE",
		},
		{
			name:   "a doujin work names its circle",
			labels: []catWorkLabel{{DisplayName: "たぬきそふと", Role: "circle"}},
			want:   "たぬきそふと",
		},
		{
			name: "no known role falls back to the first named company",
			labels: []catWorkLabel{
				{DisplayName: "", Role: "developer"},
				{DisplayName: "Kun", Role: "wat"},
			},
			want: "Kun",
		},
		{name: "no companies", labels: nil, want: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := makerName(ctx, c.labels); got != c.want {
				t.Fatalf("makerName = %q, want %q", got, c.want)
			}
		})
	}
}

func TestCatalogItemToNextMoeItem_CarriesPortraitAndMaker(t *testing.T) {
	raw := []byte(`{"object":"work","id":"7","display_name":"Kun","cover":{"url":"https://cdn.example/p.webp","width":600,"height":800,"thumbhash":"P"},"banner":{"url":"https://cdn.example/b.webp","width":1280,"height":720,"thumbhash":"B"},"companies":[{"object":"company","id":"1","display_name":"Key","company_kind":"publisher","attribution_role":"developer"}]}`)
	var it CatalogWorkListItem
	if err := json.Unmarshal(rewriteV2JSON(raw, ""), &it); err != nil {
		t.Fatalf("decode: %v", err)
	}
	m := CatalogItemToNextMoeItem(context.Background(), &it)
	if m.EffectivePortraitURL != "https://cdn.example/p.webp" {
		t.Fatalf("portrait = %q, want the portrait slot", m.EffectivePortraitURL)
	}
	if m.EffectiveBannerURL != "https://cdn.example/b.webp" {
		t.Fatalf("banner = %q, want the banner slot, not the flat cover fallback", m.EffectiveBannerURL)
	}
	if m.Company != "Key" {
		t.Fatalf("company = %q", m.Company)
	}
}
