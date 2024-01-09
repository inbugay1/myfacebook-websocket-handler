package ws

type serverMessage struct {
	PostID   string `json:"post_id"`
	PostText string `json:"post_text"`
	AuthorID string `json:"author_id"`
}
