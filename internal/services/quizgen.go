// internal/services/quizgen.go

package services

import (
    "AppDevelopmentAPI/internal/models"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"

    "github.com/google/uuid"
)

type QuizGen struct{ Key string }

func NewQuizGen() QuizGen { return QuizGen{Key: os.Getenv("OPENAI_API_KEY")} }

func (g QuizGen) Generate(name string, lat, lon float64) ([]models.Question, error) {
    body := map[string]any{
        "model": "gpt-3.5-turbo",
        "messages": []map[string]string{
            {"role": "system", "content": "You are a quiz generator."},
            {"role": "user", "content": generatePrompt(name, lat, lon)},
        },
        "temperature": 0.7,
        "max_tokens":  1000,
    }

    reqBytes, _ := json.Marshal(body)
    req, _ := http.NewRequest(
        "POST",
        "https://api.openai.com/v1/chat/completions",
        bytes.NewBuffer(reqBytes),
    )
    req.Header.Set("Authorization", "Bearer "+g.Key)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var api struct {
        Choices []struct {
            Message struct{ Content string } `json:"message"`
        } `json:"choices"`
    }
    if err = json.NewDecoder(resp.Body).Decode(&api); err != nil {
        return nil, err
    }

    var qs []models.Question
    if err = json.Unmarshal([]byte(api.Choices[0].Message.Content), &qs); err != nil {
        return nil, err
    }
    for i := range qs {
        qs[i].ID = uuid.NewString()
    }
    return qs, nil
}

func generatePrompt(name string, lat, lon float64) string {
    return fmt.Sprintf(
        `Generate 7 JSON questions (array) about "%s" (%.5f,%.5f). Each item must be {"text":...,"options":[4],"answer":number}. No markdown.`,
        name, lat, lon,
    )
}
