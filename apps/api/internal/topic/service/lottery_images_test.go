package service

import (
	"strings"
	"testing"

	topicModel "kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/imageclient"
)

const testCDN = "https://img.example.com"

func hash(c string) string { return strings.Repeat(c, 64) }

func TestPrizeImageURLsKeepsPositionsForSFWReader(t *testing.T) {
	hashes := []string{hash("a"), hash("b"), hash("c")}
	urls := prizeImageURLs(testCDN, hashes, []string{hash("b")}, nil, true)

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
	hashes := []string{hash("a"), hash("b")}
	urls := prizeImageURLs(testCDN, hashes, []string{hash("a")}, []string{hash("b")}, false)

	for i, url := range urls {
		if url == "" {
			t.Errorf("image %d withheld from a reader who opted in", i)
		}
	}
}

func TestPrizeImageURLsWithNothingMarked(t *testing.T) {
	hashes := []string{hash("a"), hash("b")}
	for _, isSFW := range []bool{true, false} {
		for i, url := range prizeImageURLs(testCDN, hashes, nil, nil, isSFW) {
			if url == "" {
				t.Errorf("isSFW=%v: image %d withheld with nothing marked", isSFW, i)
			}
		}
	}
}

func TestPrizeImageURLsWithholdsWhatOnlyTheGraderMarked(t *testing.T) {
	hashes := []string{hash("a"), hash("b")}
	urls := prizeImageURLs(testCDN, hashes, nil, []string{hash("b")}, true)

	if urls[0] == "" {
		t.Errorf("clean image withheld")
	}
	if urls[1] != "" {
		t.Errorf("an image the author never marked but the grader called explicit must still be withheld, got %q", urls[1])
	}
}

func TestMachineExplicitHashesNeedsAResolver(t *testing.T) {
	s := &LotteryService{}
	if got := s.machineExplicitHashes(prizesWithHashes(hash("a"))); got != nil {
		t.Errorf("without an image client nothing may be machine marked, got %v", got)
	}
}

func TestMachineExplicitHashesTakesOnlyTheExplicitOnes(t *testing.T) {
	level := func(v int16) *int16 { return &v }
	s := &LotteryService{imageMeta: func([]string) map[string]imageclient.ImageMeta {
		return map[string]imageclient.ImageMeta{
			hash("a"): {Sexual: level(0)},
			hash("b"): {Sexual: level(1)},
			hash("c"): {Sexual: level(2)},
			hash("d"): {}, // ungraded: the grader has not reached it yet
		}
	}}

	got := s.machineExplicitHashes(prizesWithHashes(hash("a"), hash("b"), hash("c"), hash("d")))

	if len(got) != 1 || !got[hash("c")] {
		t.Errorf("only an explicit grade may mark an image, got %v", got)
	}
}

func prizesWithHashes(hashes ...string) []topicModel.TopicLotteryPrize {
	return []topicModel.TopicLotteryPrize{{ImageHashes: hashes}}
}
