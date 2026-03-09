package mapper

import (
	"fmt"
	"notification-service/internal/domain"
	"strings"
)

func ConvertPostToNotification(post *domain.Post) string {
	res := fmt.Sprintf("User %s created new post with title %s and tags %s", post.Author, post.Title, strings.Join(post.Tags, ","))
	return res
}
