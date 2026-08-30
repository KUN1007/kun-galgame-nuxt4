package service

import (
	"strings"
	"testing"
)

const testCDN = "https://img.example.com"

func TestPrizeImageURLsKeepsPositionsForSFWReader(t *testing.T) {
	hashes := []string{strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)}
	urls := prizeImageURLs(testCDN, hashes, []string{strings.Repeat("b", 64)}, true)

	if len(urls) != len(hashes) {
		t.Fatalf("want %d urls, got %d", len(hashes), len(urls))
	}
	if urls[1] != "" {
		t.Errorf("marked image should have no url, got %q", urls[1])
	}
	if urls[0] == "" || urls[2] == "" {
		t.Errorf("unmarked images must stay resolvable, got %q and %q", urls[0], urls[2])
	}
}

func TestPrizeImageURLsResolvesEverythingForNSFWReader(t *testing.T) {
	hashes := []string{strings.Repeat("a", 64), strings.Repeat("b", 64)}
	urls := prizeImageURLs(testCDN, hashes, []string{strings.Repeat("a", 64), strings.Repeat("b", 64)}, false)

	for i, url := range urls {
		if url == "" {
			t.Errorf("image %d withheld from a reader who opted in", i)
		}
	}
}

func TestPrizeImageURLsWithNothingMarked(t *testing.T) {
	hashes := []string{strings.Repeat("a", 64), strings.Repeat("b", 64)}
	for _, isSFW := range []bool{true, false} {
		for i, url := range prizeImageURLs(testCDN, hashes, nil, isSFW) {
			if url == "" {
				t.Errorf("isSFW=%v: image %d withheld with nothing marked", isSFW, i)
			}
		}
	}
}
