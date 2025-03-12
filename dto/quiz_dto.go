package dto

// ChoiceRequest โครงสร้างข้อมูลสำหรับตัวเลือก
type ChoiceRequest struct {
	ID        uint   `json:"id,omitempty"`
	Text      string `json:"text"`
	ImageURL  string `json:"imageURL,omitempty"`
	IsCorrect bool   `json:"isCorrect"`
}

// QuestionRequest โครงสร้างข้อมูลสำหรับคำถาม
type QuestionRequest struct {
	ID       uint           `json:"id,omitempty"`
	Text     string         `json:"text"`
	ImageURL string         `json:"imageURL,omitempty"`
	Choices  []ChoiceRequest `json:"choices"`
}

// UpdateQuizRequest โครงสร้างข้อมูลสำหรับการอัปเดต quiz
type UpdateQuizRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TimeLimit   uint              `json:"timeLimit"`
	IsPublished bool              `json:"isPublished"`
	Categories  []uint            `json:"categories"`
	Questions   []QuestionRequest `json:"questions"`
}

type ChoiceFormData struct {
	ID        uint   `json:"Id,omitempty"`
	Text      string `json:"Text"`
	IsCorrect bool   `json:"IsCorrect"`
}

// QuestionFormRequest สำหรับคำถามใน FormData
type QuestionFormData struct {
	ID      uint               `json:"Id,omitempty"`
	Text    string             `json:"Text"`
	Choices []ChoiceFormData `json:"Choices"`
}

// QuizFormData สำหรับฟอร์มข้อมูล quiz
type QuizFormData struct {
	Title       string                `json:"Title"`
	Description string                `json:"Description"`
	TimeLimit   uint                  `json:"TimeLimit"`
	IsPublished bool                  `json:"IsPublished"`
	Categories  []uint                `json:"Categories"`
	Questions   []QuestionFormData `json:"Questions"`
}

// CreateQuestionRequest โครงสร้างข้อมูลสำหรับการสร้างคำถาม
type CreateQuestionRequest struct {
	Text     string            `json:"text"`
	ImageURL string            `json:"imageURL,omitempty"`
	Choices  []CreateChoiceRequest `json:"choices"`
}

// CreateChoiceRequest โครงสร้างข้อมูลสำหรับการสร้างตัวเลือก
type CreateChoiceRequest struct {
	Text      string `json:"text"`
	ImageURL  string `json:"imageURL,omitempty"`
	IsCorrect bool   `json:"isCorrect"`
}
// CreateQuizRequest โครงสร้างข้อมูลสำหรับการสร้าง quiz
type CreateQuizRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TimeLimit   uint              `json:"timeLimit"`
	IsPublished bool              `json:"isPublished"`
	ImageURL    string            `json:"imageURL,omitempty"`
	Categories  []uint            `json:"categories"`
	Questions   []CreateQuestionRequest `json:"questions"`
}