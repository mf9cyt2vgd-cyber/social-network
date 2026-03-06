package utils

import (
	"encoding/json"
	"notifications-service/internal/domain"
)

func ConvertKafkaMessageIntoPost(msg []byte) (domain.Post, error) {
	var post domain.Post
	err := json.Unmarshal(msg, &post)
	if err != nil {
		return domain.Post{}, err
	}
	return post, nil
}
