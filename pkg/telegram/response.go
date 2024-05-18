package telegram

type SendMessageResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int         `json:"message_id"`
		From      interface{} `json:"from"`
		Chat      interface{} `json:"chat"`
		Date      int         `json:"date"`
		Text      string      `json:"text"`
	} `json:"result"`
}
