package mapper

import (
	"fmt"
	"notification-service/internal/domain"
	"strings"
)

func ConvertPostToNotification(post *domain.Post) string {
	res := fmt.Sprintf("[%s] User %s created new post with title %s and tags %s", post.ID, post.Author, post.Title, strings.Join(post.Tags, ","))
	return res
}
