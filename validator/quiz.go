package validator

import (
	"strings"

	model "github.com/patiphanak/league-of-quiz/model"
)

// Error struct to hold validation errors
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateQuiz validates a quiz and its associated questions and choices
func ValidateQuiz(quiz model.Quiz) []Error {
	var errors []Error

	// Validate quiz fields
	if strings.TrimSpace(quiz.Title) == "" {
		errors = append(errors, Error{Field: "title", Message: "Title is required"})
	}

	if strings.TrimSpace(quiz.Description) == "" {
		errors = append(errors, Error{Field: "description", Message: "Description is required"})
	}

	if quiz.TimeLimit == 0 {
		errors = append(errors, Error{Field: "timeLimit", Message: "Time limit must be greater than 0"})
	}

	if quiz.CreatorID == 0 {
		errors = append(errors, Error{Field: "creatorID", Message: "Creator ID is required"})
	}

	// Validate questions
	if len(quiz.Questions) == 0 {
		errors = append(errors, Error{Field: "questions", Message: "Quiz must have at least one question"})
	} else {
		for i, question := range quiz.Questions {
			questionErrors := ValidateQuestion(question)
			for _, err := range questionErrors {
				errors = append(errors, Error{
					Field:   "questions[" + string(rune(i)) + "]." + err.Field,
					Message: err.Message,
				})
			}
		}
	}

	return errors
}

// ValidateQuestion validates a question and its associated choices
func ValidateQuestion(question model.Question) []Error {
	var errors []Error

	// Validate question fields
	if strings.TrimSpace(question.Text) == "" {
		errors = append(errors, Error{Field: "text", Message: "Question text is required"})
	}

	// Validate choices
	if len(question.Choices) < 2 {
		errors = append(errors, Error{Field: "choices", Message: "Question must have at least two choices"})
	} else {
		hasCorrectChoice := false
		for i, choice := range question.Choices {
			if choice.IsCorrect {
				hasCorrectChoice = true
			}

			choiceErrors := ValidateChoice(choice)
			for _, err := range choiceErrors {
				errors = append(errors, Error{
					Field:   "choices[" + string(rune(i)) + "]." + err.Field,
					Message: err.Message,
				})
			}
		}

		if !hasCorrectChoice {
			errors = append(errors, Error{Field: "choices", Message: "Question must have at least one correct choice"})
		}
	}

	return errors
}

// ValidateChoice validates a choice
func ValidateChoice(choice model.Choice) []Error {
	var errors []Error

	if strings.TrimSpace(choice.Text) == "" {
		errors = append(errors, Error{Field: "text", Message: "Choice text is required"})
	}

	return errors
}

// ValidateGameSession validates a game session
func ValidateGameSession(session model.GameSession) []Error {
	var errors []Error

	if session.QuizID == 0 {
		errors = append(errors, Error{Field: "quizID", Message: "Quiz ID is required"})
	}

	if session.HostID == 0 {
		errors = append(errors, Error{Field: "hostID", Message: "Host ID is required"})
	}

	// Validate status
	status := strings.ToLower(session.Status)
	if status != "waiting" && status != "playing" && status != "finished" {
		errors = append(errors, Error{Field: "status", Message: "Status must be 'waiting', 'playing', or 'finished'"})
	}

	return errors
}

// ValidateGamePlayer validates a game player
func ValidateGamePlayer(player model.GamePlayer) []Error {
	var errors []Error

	if strings.TrimSpace(player.SessionID) == "" {
		errors = append(errors, Error{Field: "sessionID", Message: "Session ID is required"})
	}

	if player.UserID == 0 {
		errors = append(errors, Error{Field: "userID", Message: "User ID is required"})
	}

	if strings.TrimSpace(player.Nickname) == "" {
		errors = append(errors, Error{Field: "nickname", Message: "Nickname is required"})
	}

	return errors
}

// ValidatePlayerAnswer validates a player answer
func ValidatePlayerAnswer(answer model.PlayerAnswer) []Error {
	var errors []Error

	if strings.TrimSpace(answer.SessionID) == "" {
		errors = append(errors, Error{Field: "sessionID", Message: "Session ID is required"})
	}

	if answer.QuizID == 0 {
		errors = append(errors, Error{Field: "quizID", Message: "Quiz ID is required"})
	}

	if answer.QuestionID == 0 {
		errors = append(errors, Error{Field: "questionID", Message: "Question ID is required"})
	}

	if answer.PlayerID == 0 {
		errors = append(errors, Error{Field: "playerID", Message: "Player ID is required"})
	}

	if answer.ChoiceID == 0 {
		errors = append(errors, Error{Field: "choiceID", Message: "Choice ID is required"})
	}

	if answer.TimeSpent < 0 {
		errors = append(errors, Error{Field: "timeSpent", Message: "Time spent must be a positive value"})
	}

	return errors
}
