package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// Relevance weights. A title hit outranks a category hit outranks a body hit,
// and the same words sitting next to each other outrank them scattered.
const (
	weightTitle         = 8
	weightCategory      = 3
	weightContent       = 1
	weightPhraseTitle   = 20
	weightPhraseContent = 10
	// One long rant repeating the term must not outrank everything else.
	maxRepeatBonus = 5

	snippetLen  = 233
	snippetLead = 30
)

// adjacencyPattern matches the keywords sitting next to each other whatever
// separator the prose uses. Matching the raw query verbatim does not work here:
// a search typed "汉化 补丁" has to score the contiguous "汉化补丁" that Chinese
// actually writes, which carries no space at all.
//
// Empty when there is nothing to discriminate — with a single keyword every
// matched row contains the phrase, so the bonus would be a constant added to
// the entire result set.
func adjacencyPattern(keywords []string) string {
	if len(keywords) < 2 {
		return ""
	}
	quoted := make([]string, len(keywords))
	for i, kw := range keywords {
		quoted[i] = regexp.QuoteMeta(kw)
	}
	return strings.Join(quoted, `\s*`)
}

// topicRelevance scores a topic row by which field each keyword landed in.
func topicRelevance(keywords []string) (string, []any) {
	parts := []string{"0"}
	args := make([]any, 0, len(keywords)*3+2)
	for _, kw := range keywords {
		like := "%" + kw + "%"
		parts = append(parts,
			fmt.Sprintf("(CASE WHEN t.title ILIKE ? THEN %d ELSE 0 END)", weightTitle),
			fmt.Sprintf("(CASE WHEN t.category ILIKE ? THEN %d ELSE 0 END)", weightCategory),
			fmt.Sprintf("(CASE WHEN t.content ILIKE ? THEN %d ELSE 0 END)", weightContent),
		)
		args = append(args, like, like, like)
	}
	if phrase := adjacencyPattern(keywords); phrase != "" {
		parts = append(parts,
			fmt.Sprintf("(CASE WHEN t.title ~* ? THEN %d ELSE 0 END)", weightPhraseTitle),
			fmt.Sprintf("(CASE WHEN t.content ~* ? THEN %d ELSE 0 END)", weightPhraseContent),
		)
		args = append(args, phrase, phrase)
	}
	return strings.Join(parts, " + "), args
}

// contentRelevance scores a body-only row (replies, comments). Every keyword is
// already present because the filter ANDs them, so what is left to discriminate
// is adjacency and how often the leading term repeats.
func contentRelevance(col string, keywords []string) (string, []any) {
	parts := []string{"0"}
	var args []any
	if phrase := adjacencyPattern(keywords); phrase != "" {
		parts = append(parts,
			fmt.Sprintf("(CASE WHEN %s ~* ? THEN %d ELSE 0 END)", col, weightPhraseContent))
		args = append(args, phrase)
	}
	lead := keywords[0]
	parts = append(parts, fmt.Sprintf(
		"LEAST(%d, (length(lower(%s)) - length(replace(lower(%s), ?, ''))) / %d)",
		maxRepeatBonus, col, col, len([]rune(lead))))
	args = append(args, strings.ToLower(lead))
	return strings.Join(parts, " + "), args
}

// contentSnippet windows the body around the earliest keyword hit. The old
// excerpt was always the first 233 characters, so a hit deep in a long reply
// produced a preview with none of the searched words in it.
func contentSnippet(col string, keywords []string) (string, []any) {
	hits := make([]string, 0, len(keywords))
	args := make([]any, 0, len(keywords))
	for _, kw := range keywords {
		hits = append(hits, fmt.Sprintf("NULLIF(POSITION(? IN lower(%s)), 0)", col))
		args = append(args, strings.ToLower(kw))
	}
	first := hits[0]
	if len(hits) > 1 {
		// LEAST skips NULLs, so a keyword that only matched another column
		// does not drag the window back to the start.
		first = "LEAST(" + strings.Join(hits, ", ") + ")"
	}
	expr := fmt.Sprintf(
		"CASE WHEN %s > %d THEN '…' || SUBSTRING(%s FROM %s - %d FOR %d) ELSE SUBSTRING(%s FROM 1 FOR %d) END",
		first, snippetLead, col, first, snippetLead, snippetLen, col, snippetLen)
	// `first` is emitted twice, so its placeholders are bound twice in order.
	return expr, append(args, args...)
}
