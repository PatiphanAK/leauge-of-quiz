package dto

type QuestionChoiceForm struct {
	ID        string  `json:"id,omitempty"` 
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}
type QuestionFormData struct {
	QuizID  int             `json:"quizId"`
	Text    string          `json:"text"`
	Choices []ChoiceFormData `json:"choices"`
}

// ChoiceFormData สำหรับข้อมูลตัวเลือกจาก form
type ChoiceFormData struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}