package main

import "encoding/json"
import "log"
import "fmt"

var subjectsCache *Subjects
var subjectsDataMap map[int]SubjectsData = make(map[int]SubjectsData)

type Subjects struct {
    Data []SubjectsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int64  `json:"total_count"`
    URL        string `json:"url"`
}

type SubjectsData struct {
    Data struct {
        Character       string `json:"character"`
        CharacterImages []struct {
            ContentType string `json:"content_type"`
            URL         string `json:"url"`
        } `json:"character_images"`
        Characters          string  `json:"characters"`
        ComponentSubjectIds []int `json:"component_subject_ids"`
        CreatedAt           string  `json:"created_at"`
        DocumentURL         string  `json:"document_url"`
        Level               int64   `json:"level"`
        Meanings            []struct {
            Meaning string `json:"meaning"`
            Primary bool   `json:"primary"`
        } `json:"meanings"`
        PartsOfSpeech []string `json:"parts_of_speech"`
        Readings      []struct {
            Primary bool   `json:"primary"`
            Reading string `json:"reading"`
            Type    string `json:"type"`
        } `json:"readings"`
        Slug string `json:"slug"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    ID            int  `json:"id"`
    Object        string `json:"object"`
    URL           string `json:"url"`
} 

func getSubjects(apiKey string, chResult chan *Subjects) {
    if subjectsCache != nil {
        chResult <- subjectsCache
        return
    }

    ch := make(chan *Subjects)

    maxPages := 18
    for page := 1; page <= maxPages; page++ {
        go getSubjectsPage(apiKey, page, ch)
    }
    
    subjects := <-ch
    if (int(subjects.Pages.Last) > maxPages) {
        for page := maxPages+1; page <= int(subjects.Pages.Last); page++ {
            go getSubjectsPage(apiKey, page, ch)
        }
        maxPages = int(subjects.Pages.Last)
    }

    for page := 2; page <= maxPages; page++ {
        subjectsPage := <-ch
        subjects.Data = append(subjects.Data, subjectsPage.Data...)
    }

    subjects.Pages.Current = 1

    for i := 0; i<len(subjects.Data); i++ {
        subjectsDataMap[subjects.Data[i].ID] = subjects.Data[i]
    }
    subjectsCache = subjects
    
    chResult <- subjects
}

func getSubjectsPage(apiKey string, page int, ch chan *Subjects) {
    body := getUrl(apiKey, fmt.Sprintf("https://wanikani.com/api/v2/subjects?page=%d",page))
    var subjects Subjects
    
    err := json.Unmarshal(body, &subjects)
    if err != nil {
        log.Fatal("error:", err, string(body))
        panic(err)
    }

    ch <- &subjects
}