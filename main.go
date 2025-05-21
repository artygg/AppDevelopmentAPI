package main

import (
    "bytes"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "github.com/joho/godotenv"
)

type Place struct {
    Name       string     `json:"name"`
    Coordinate Coordinate `json:"coordinate"`
    CategoryID int        `json:"category_id"`
}
type Coordinate struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}
type FoursquareResponse struct {
    Results []FoursquarePlace `json:"results"`
}
type FoursquarePlace struct {
    FsqID      string          `json:"fsq_id"`
    Name       string          `json:"name"`
    Categories []FoursquareCat `json:"categories"`
    Geocodes   struct {
        Main struct {
            Latitude  float64 `json:"latitude"`
            Longitude float64 `json:"longitude"`
        } `json:"main"`
    } `json:"geocodes"`
}
type FoursquareCat struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
type Question struct {
    Text    string   `json:"text"`
    Options []string `json:"options"`
    Answer  int      `json:"answer"` // index of correct answer
}
type Quiz struct {
    PlaceID   string     `json:"place_id"`
    Questions []Question `json:"questions"`
}

func loadAllowedCategoryIDs(path string) (map[int]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    r := csv.NewReader(file)
    _, err = r.Read()
    if err != nil {
        return nil, err
    }
    ids := make(map[int]string)
    for {
        record, err := r.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        id, err := strconv.Atoi(record[0])
        if err != nil {
            continue
        }
        ids[id] = record[1]
    }
    return ids, nil
}

func fetchAndSavePOIs(apiKey string, centerLat, centerLon float64, allowedCategoryIDs map[int]string) error {
    const (
        radius    = 50000 // 50km
        pageLimit = 50
    )
    var places []Place
    dedupSet := make(map[string]struct{})
    cursor := ""
    client := &http.Client{}
    for {
        url := fmt.Sprintf("https://api.foursquare.com/v3/places/search?ll=%.6f,%.6f&radius=%d&limit=%d", centerLat, centerLon, radius, pageLimit)
        if cursor != "" {
            url += "&cursor=" + cursor
        }
        req, _ := http.NewRequest("GET", url, nil)
        req.Header.Set("Authorization", apiKey)
        req.Header.Set("Accept", "application/json")

        resp, err := client.Do(req)
        if err != nil {
            return err
        }
        if resp.StatusCode != 200 {
            body, _ := io.ReadAll(resp.Body)
            resp.Body.Close()
            return fmt.Errorf("API error: %s\n%s", resp.Status, string(body))
        }
        var fsqResp struct {
            Results []FoursquarePlace `json:"results"`
            Context struct {
                NextCursor string `json:"next_cursor"`
            } `json:"context"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&fsqResp); err != nil {
            resp.Body.Close()
            return err
        }
        resp.Body.Close()

        for _, fsq := range fsqResp.Results {
            matchedCategoryID := 0
            for _, cat := range fsq.Categories {
                if _, ok := allowedCategoryIDs[cat.ID]; ok {
                    matchedCategoryID = cat.ID // Save the first allowed category ID
                    break
                }
            }
            if matchedCategoryID == 0 {
                continue
            }
            key := fmt.Sprintf("%s|%.6f|%.6f", fsq.Name, fsq.Geocodes.Main.Latitude, fsq.Geocodes.Main.Longitude)
            if _, exists := dedupSet[key]; exists {
                continue
            }
            dedupSet[key] = struct{}{}
            places = append(places, Place{
                Name: fsq.Name,
                Coordinate: Coordinate{
                    Latitude:  fsq.Geocodes.Main.Latitude,
                    Longitude: fsq.Geocodes.Main.Longitude,
                },
                CategoryID: matchedCategoryID,
            })
        }
        if fsqResp.Context.NextCursor == "" {
            break
        }
        cursor = fsqResp.Context.NextCursor
    }

    f, err := os.Create("places.json")
    if err != nil {
        return err
    }
    defer f.Close()
    data, err := json.MarshalIndent(places, "", "  ")
    if err != nil {
        return err
    }
    _, err = f.Write(data)
    return err
}

func generateQuizForPlace(placeName, apiKey string) (Quiz, error) {
   prompt := fmt.Sprintf(`
    You are generating a quiz for a location-based game.
    - Generate exactly 7 questions.
    - The first 3 questions must be about "%s" (city, facts, what it's known for, location, etc).
    - Each question must be a valid object with:
      - a non-empty "text" field (the question, in English)
      - an "options" field: 4 possible answers
      - an "answer" field: the 0-based index of the correct answer in "options".
    - NEVER use the field "question", only "text".
    - Do not use markdown or explanation, just respond with a JSON array:
    [
      {"text":"What is the capital of France?","options":["Paris","Berlin","Rome","Madrid"],"answer":0}
    ]
    If any "text" value is empty, REGENERATE THE QUESTION with a real question in English.
    Return ONLY valid JSON, not markdown.
    `, placeName)


    requestBody := map[string]interface{}{
        "model": "gpt-3.5-turbo",
        "messages": []map[string]string{
            {"role": "system", "content": "You are a quiz generator for a location-based game."},
            {"role": "user", "content": prompt},
        },
        "temperature": 0.7,
        "max_tokens": 1000,
    }
    reqBytes, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBytes))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return Quiz{}, err
    }
    defer resp.Body.Close()
    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return Quiz{}, err
    }

    var apiResp struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    if err := json.Unmarshal(respBody, &apiResp); err != nil {
        return Quiz{}, err
    }

    var questions []Question
    cleaned := apiResp.Choices[0].Message.Content
    cleanedBytes := []byte(cleaned)
    cleanedBytes = bytes.TrimPrefix(cleanedBytes, []byte("```json\n"))
    cleanedBytes = bytes.TrimSuffix(cleanedBytes, []byte("\n```"))
    if err := json.Unmarshal(cleanedBytes, &questions); err != nil {
        if err := json.Unmarshal([]byte(apiResp.Choices[0].Message.Content), &questions); err != nil {
            return Quiz{}, err
        }
    }

    return Quiz{
        PlaceID:   placeName,
        Questions: questions,
    }, nil
}

func quizFilePath(place string) string {
    safe := strings.ReplaceAll(place, " ", "_")
    return filepath.Join("quizzes", safe + ".json")
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    apiKey := os.Getenv("FSQ_API_KEY")
    openaiKey := os.Getenv("OPENAI_API_KEY")
    centerLatStr := os.Getenv("FSQ_CENTER_LAT")
    centerLonStr := os.Getenv("FSQ_CENTER_LON")
    centerLat, _ := strconv.ParseFloat(centerLatStr, 64)
    centerLon, _ := strconv.ParseFloat(centerLonStr, 64)

    allowedCategoryIDs, err := loadAllowedCategoryIDs("poi_interesting_categories.csv")
    if err != nil {
        log.Fatalf("Could not load allowed categories: %v", err)
    }

    err = fetchAndSavePOIs(apiKey, centerLat, centerLon, allowedCategoryIDs)
    if err != nil {
        log.Fatalf("Failed to fetch/save POIs: %v", err)
    }

    // --- /places ---
    http.HandleFunc("/places", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        f, err := os.Open("places.json")
        if err != nil {
            http.Error(w, "No places found", http.StatusInternalServerError)
            return
        }
        defer f.Close()
        io.Copy(w, f)
    })

    // --- /quiz ---
    http.HandleFunc("/quiz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        place := r.URL.Query().Get("place")
        if place == "" {
            http.Error(w, "Missing place parameter", http.StatusBadRequest)
            return
        }
        quizPath := quizFilePath(place)
        if f, err := os.Open(quizPath); err == nil {
            defer f.Close()
            io.Copy(w, f)
            return
        }
        quiz, err := generateQuizForPlace(place, openaiKey)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to generate quiz: %v", err), http.StatusInternalServerError)
            return
        }
        // Save to file for next time
        if err := os.MkdirAll("quizzes", 0755); err == nil { // Ensure folder exists
            if f, err := os.Create(quizPath); err == nil {
                json.NewEncoder(f).Encode(quiz)
                f.Close()
            }
        }
        json.NewEncoder(w).Encode(quiz)
    })

    http.Handle("/", http.FileServer(http.Dir(".")))

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
