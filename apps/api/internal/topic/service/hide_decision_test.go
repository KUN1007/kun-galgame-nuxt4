package service

import (
	"testing"

	topicModel "kun-galgame-api/internal/topic/model"
)

func TestHideDecision(t *testing.T) {
	tests := []struct {
		name      string
		topic     topicModel.Topic
		uid       int
		can       bool
		status    int
		by        string
		forbidden bool
	}{
		{"author hide", topicModel.Topic{UserID: 1}, 1, false, 1, "author", false},
		{"moderator hide", topicModel.Topic{UserID: 1}, 2, true, 1, "moderator", false},
		{"other hide", topicModel.Topic{UserID: 1}, 2, false, 0, "", true},
		{"author unhide", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "author"}, 1, false, 0, "", false},
		{"author moderator refusal", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "moderator"}, 1, false, 1, "moderator", true},
		{"author trust refusal", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "trust"}, 1, false, 1, "trust", true},
		{"moderator-author unhide", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "moderator"}, 1, true, 0, "", false},
		{"moderator unhide", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "trust"}, 2, true, 0, "", false},
		{"other unhide", topicModel.Topic{UserID: 1, Status: 1, HiddenBy: "author"}, 2, false, 1, "author", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, b, e := hideDecision(&tt.topic, tt.uid, tt.can)
			if s != tt.status || b != tt.by || (e != nil) != tt.forbidden {
				t.Fatalf("got %d %q %v", s, b, e)
			}
		})
	}
}
