package gate

import "kun-galgame-api/pkg/trustclient"

const (
	SubjectKindGalgameComment = "galgame_comment"
	SubjectKindGalgame        = "galgame"
	SubjectKindUser           = "user"
	SubjectKindTodo           = "forum_todo"
)

var CanonicalSubjectKinds = []string{
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

func CanonicalSubjectKindItems() []trustclient.EnsureSubjectKindItem {
	items := make([]trustclient.EnsureSubjectKindItem, len(CanonicalSubjectKinds))
	for i, k := range CanonicalSubjectKinds {
		items[i] = trustclient.EnsureSubjectKindItem{Key: k}
	}
	return items
}
