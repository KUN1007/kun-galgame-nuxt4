package markdown

import (
	"strings"
	"testing"
)

func TestRenderStripsXSS(t *testing.T) {
	cases := []struct {
		name   string
		md     string
		banned []string
	}{
		{"script", "hi <script>alert(1)</script> there", []string{"<script", "alert(1)"}},
		{"img onerror", "<img src=x onerror=alert(1)>", []string{"onerror"}},
		{"js link", "[click](javascript:alert(1))", []string{"javascript:"}},
		{"iframe", "<iframe src=//evil.com></iframe>", []string{"<iframe"}},
		{"onclick", `<p onclick="x()">p</p>`, []string{"onclick"}},
		{"style tag", "<style>*{x:y}</style>t", []string{"<style"}},
		{"svg onload", `<svg onload=alert(1)>`, []string{"onload"}},
		{"spoiler xss", "||<script>bad()</script>||", []string{"<script", "bad()"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
			for _, b := range c.banned {
				if strings.Contains(out, b) {
					t.Errorf("must NOT contain %q\n got: %s", b, out)
				}
			}
		})
	}
}

func TestRenderPreservesFeatures(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want []string
	}{
		{"heading id", "## Hello World", []string{"<h2", "id="}},
		{"safe link", "[ok](https://example.com)", []string{"https://example.com"}},
		{"bold italic", "**b** _i_", []string{"<strong>b</strong>", "<em>i</em>"}},
		{"code block", "```go\nfmt.Println()\n```", []string{"kun-code-container", "language-go", "copy"}},
		{"code block no lang", "```\nplain text\n```", []string{"kun-code-container", "copy"}},
		{"table", "| a | b |\n|---|---|\n| 1 | 2 |", []string{"kun-table-container", "<table>"}},
		{"spoiler", "||secret||", []string{"kun-spoiler", "secret"}},
		{"video", "kv:[v](https://e.com/x.mp4)", []string{"<video", "e.com/x.mp4"}},
		{"lazy image", "![alt](https://e.com/i.png)", []string{"data-kun-lazy-image", "loading=", "e.com/i.png"}},
		{"list", "- one\n- two", []string{"<ul>", "<li>"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("must contain %q\n got: %s", w, out)
				}
			}
		})
	}
}

func TestRenderCodeBlocksDivBalanced(t *testing.T) {
	cases := map[string]string{
		"no lang":         "before\n\n```\nplain\n```\n\nafter",
		"with lang":       "before\n\n```go\nx := 1\n```\n\nafter",
		"mixed":           "```\nplain\n```\n\ntext\n\n```sql\nselect 1\n```",
		"indented":        "    indented code\n",
		"angle-bracket <": "```\nif a < b then\n```",
	}
	for name, md := range cases {
		t.Run(name, func(t *testing.T) {
			out := Render(md)
			if open, clo := strings.Count(out, "<div"), strings.Count(out, "</div>"); open != clo {
				t.Errorf("unbalanced divs: %d open vs %d close\n got: %s", open, clo, out)
			}
		})
	}
}

func TestRenderMention(t *testing.T) {
	out := Render("hi [@白狐](kungal-user:123) there")
	for _, w := range []string{`class="kun-mention"`, `data-uid="123"`, "@白狐"} {
		if !strings.Contains(out, w) {
			t.Errorf("mention must contain %q\n got: %s", w, out)
		}
	}
	if strings.Contains(out, "kungal-user:") {
		t.Errorf("raw mention token leaked the custom scheme: %s", out)
	}
}

func TestRenderQuote(t *testing.T) {
	out := Render("see [#12](kungal-reply:456)")
	for _, w := range []string{`class="kun-quote"`, `data-reply-id="456"`, `data-floor="12"`, "#12"} {
		if !strings.Contains(out, w) {
			t.Errorf("quote must contain %q\n got: %s", w, out)
		}
	}
	if strings.Contains(out, "kungal-reply:") {
		t.Errorf("raw quote token leaked the custom scheme: %s", out)
	}
}

func TestRenderMentionAbsoluteHref(t *testing.T) {
	SetContentSiteBase("https://www.kungal.com")
	defer SetContentSiteBase("")
	out := Render("[@白狐](kungal-user:123)")
	if !strings.Contains(out, `href="https://www.kungal.com/user/123/info"`) {
		t.Errorf("mention should have an absolute profile href\n got: %s", out)
	}
}

func TestExtractMentionIDs(t *testing.T) {
	got := ExtractMentionIDs(
		"hi [@a](kungal-user:5) and [@b](kungal-user:12), and [@a again](kungal-user:5)",
	)
	want := []int{5, 12}
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at %d: want %d, got %d (%v)", i, want[i], got[i], got)
		}
	}
	if ids := ExtractMentionIDs("no mentions here"); ids != nil {
		t.Errorf("expected nil for no mentions, got %v", ids)
	}
}

func TestResolveMentionNames(t *testing.T) {
	html := Render("[@旧名](kungal-user:5) and [@x](kungal-user:9)")

	out := ResolveMentionNames(html, map[int]string{5: "新名"})
	if !strings.Contains(out, "@新名") {
		t.Errorf("expected resolved @新名\n got: %s", out)
	}
	if strings.Contains(out, "@旧名") {
		t.Errorf("snapshot name should be replaced\n got: %s", out)
	}
	if !strings.Contains(out, "@x") {
		t.Errorf("unresolved id 9 should keep its snapshot @x\n got: %s", out)
	}
	if !strings.Contains(out, `data-uid="5"`) || !strings.Contains(out, `data-uid="9"`) {
		t.Errorf("data-uid (link target) must be preserved\n got: %s", out)
	}

	esc := ResolveMentionNames(Render("[@x](kungal-user:5)"), map[int]string{5: "<b>x"})
	if strings.Contains(esc, "<b>x") || !strings.Contains(esc, "&lt;b&gt;x") {
		t.Errorf("resolved name must be HTML-escaped\n got: %s", esc)
	}

	if ResolveMentionNames(html, nil) != html {
		t.Errorf("nil names should be a no-op")
	}

	mig := Render("[@](kungal-user:30) [#1](kungal-reply:14)")
	for _, w := range []string{
		`class="kun-mention"`, `data-uid="30"`,
		`class="kun-quote"`, `data-reply-id="14"`, `data-floor="1"`,
	} {
		if !strings.Contains(mig, w) {
			t.Errorf("empty-name migration token missing %q\n got: %s", w, mig)
		}
	}
	if r := ResolveMentionNames(mig, map[int]string{30: "鲲"}); !strings.Contains(r, "@鲲") {
		t.Errorf("empty-name mention should resolve to current name\n got: %s", r)
	}
}

func TestRenderQuestionPlain(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
		bad  []string
	}{
		{
			name: "plain text",
			in:   "谁是莲的母亲？",
			want: []string{"谁是莲的母亲？"},
			bad:  []string{"<", ">", "kun-spoiler"},
		},
		{
			name: "spoiler",
			in:   "||剧透内容||",
			want: []string{`class="kun-spoiler`, "kun-spoiler-hidden", "剧透内容"},
		},
		{
			name: "mixed",
			in:   "前文||剧透||后文",
			want: []string{"前文", `<span class="kun-spoiler`, "剧透", "后文"},
		},
		{
			name: "xss escaped",
			in:   "<script>alert(1)</script>",
			want: []string{"&lt;script&gt;", "alert(1)"},
			bad:  []string{"<script"},
		},
		{
			name: "spoiler across lines",
			in:   "||第一行\n第二行||",
			want: []string{`class="kun-spoiler`, "第一行\n第二行</span>"},
			bad:  []string{"||"},
		},
		{
			name: "spoiler xss escaped",
			in:   "||<img src=x onerror=alert(1)>||",
			want: []string{"kun-spoiler", "&lt;img", "onerror=alert(1)"},
			bad:  []string{"<img"},
		},
		{
			name: "no markdown parse",
			in:   "## 标题 **粗体**",
			want: []string{"## 标题 **粗体**"},
			bad:  []string{"<h", "<strong>", "<ul>"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := RenderQuestionPlain(c.in)
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("must contain %q\n got: %s", w, out)
				}
			}
			for _, b := range c.bad {
				if strings.Contains(out, b) {
					t.Errorf("must NOT contain %q\n got: %s", b, out)
				}
			}
		})
	}
}
