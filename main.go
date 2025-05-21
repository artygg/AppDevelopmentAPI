package main

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "github.com/joho/godotenv"
)

type Place struct {
    Name       string     `json:"name"`
    Coordinate Coordinate `json:"coordinate"`
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
            hasAllowed := false
            for _, cat := range fsq.Categories {
                if _, ok := allowedCategoryIDs[cat.ID]; ok {
                    hasAllowed = true
                    break
                }
            }
            if !hasAllowed {
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

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    apiKey := os.Getenv("FSQ_API_KEY")
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

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
