package markdown

import (
	"strings"
	"testing"
)

func TestRenderSpoilerInline(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want []string
		bad  []string
	}{
		{
			name: "soft break",
			md:   "||第一行\n第二行||",
			want: []string{`<span class="kun-spoiler`, "第一行\n第二行</span>"},
			bad:  []string{"||"},
		},
		{
			name: "hard break",
			md:   "||第一行  \n第二行||",
			want: []string{`<span class="kun-spoiler`, "第一行<br>\n第二行</span>"},
			bad:  []string{"||"},
		},
		{
			name: "three lines",
			md:   "||一\n二\n三||",
			want: []string{"一\n二\n三</span>"},
			bad:  []string{"||"},
		},
		{
			name: "blank line splits the paragraph, so it stays literal",
			md:   "||第一行\n\n第二行||",
			want: []string{"<p>||第一行</p>", "<p>第二行||</p>"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "padding spaces are kept",
			md:   "|| 带空格 ||",
			want: []string{"> 带空格 </span>"},
		},
		{
			name: "inline markup inside",
			md:   "||**粗体**||",
			want: []string{"<strong>粗体</strong></span>"},
		},
		{
			name: "code block is not spoiler syntax",
			md:   "```\n||代码块里的||\n```",
			want: []string{"<code>||代码块里的||"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "code span is not spoiler syntax",
			md:   "`||行内代码||`",
			want: []string{"<code>||行内代码||</code>"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "unpaired runs stay literal",
			md:   "好长的文||||||||||||",
			want: []string{"好长的文||||||||||||"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "table cells are untouched",
			md:   "| a | b |\n| - | - |\n| c | d |",
			want: []string{"<td>c</td>", "<td>d</td>"},
			bad:  []string{"kun-spoiler"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
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

func TestRenderSpoilerBlock(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want []string
		bad  []string
	}{
		{
			name: "wraps several paragraphs",
			md:   ":::spoiler\n第一段\n\n第二段\n:::",
			want: []string{`<div class="kun-spoiler`, "<p>第一段</p>", "<p>第二段</p>", "</div>"},
			bad:  []string{":::"},
		},
		{
			name: "block children keep their own rendering",
			md:   ":::spoiler\n- 一\n- 二\n\n```go\nx()\n```\n:::",
			want: []string{`<div class="kun-spoiler`, "<ul>", "<li>一</li>", "kun-code-container", "language-go"},
		},
		{
			name: "interrupts and resumes surrounding text",
			md:   "前\n\n:::spoiler\n中\n:::\n\n后",
			want: []string{"<p>前</p>", `<div class="kun-spoiler`, "<p>中</p>", "</div>", "<p>后</p>"},
		},
		{
			name: "unclosed fence runs to the end",
			md:   ":::spoiler\n没有闭合",
			want: []string{`<div class="kun-spoiler`, "<p>没有闭合</p>", "</div>"},
		},
		{
			name: "a longer fence nests a shorter one",
			md:   "::::spoiler\n外\n\n:::spoiler\n内\n:::\n::::",
			want: []string{"<p>外</p>", "<p>内</p>"},
			bad:  []string{":::"},
		},
		{
			name: "a code fence keeps the fence literal",
			md:   "```\n:::spoiler\nx\n:::\n```",
			want: []string{"<code>:::spoiler"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "works inside a blockquote",
			md:   "> :::spoiler\n> 引用里的\n> :::",
			want: []string{"<blockquote>", `<div class="kun-spoiler`, "<p>引用里的</p>"},
		},
		{
			name: "works inside a list item",
			md:   "- :::spoiler\n  列表里的\n  :::",
			want: []string{"<li>", `<div class="kun-spoiler`, "<p>列表里的</p>"},
		},
		{
			name: "four spaces is an indented code block, not a fence",
			md:   "    :::spoiler\n    x\n    :::",
			want: []string{"<code>:::spoiler"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "the name must touch the colons",
			md:   "::: spoiler\nx\n:::",
			want: []string{"::: spoiler"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "a trailing label is not a fence",
			md:   ":::spoiler 带标签\nx\n:::",
			want: []string{":::spoiler 带标签"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "another directive name stays literal",
			md:   ":::notspoiler\nx\n:::",
			want: []string{":::notspoiler"},
			bad:  []string{"kun-spoiler"},
		},
		{
			name: "xss inside the container is still sanitized",
			md:   ":::spoiler\n<script>bad()</script>\n:::",
			want: []string{`<div class="kun-spoiler`},
			bad:  []string{"<script", "bad()"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
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
