package main

import "encoding/json"
import "fmt"

var subjectsCache *Subjects
var subjectsDataMap map[int]SubjectsData = make(map[int]SubjectsData)

type Subjects struct {
    Data []SubjectsData `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         Pages `json:"pages"`
    TotalCount int  `json:"total_count"`
    URL        string `json:"url"`
    Error      string `json:"-"`
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
        Level               int   `json:"level"`
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

    subjects, err := getSubjectsPage(apiKey, "https://wanikani.com/api/v2/subjects")
    if err != nil {
        chResult <- &Subjects{Error: err.Error()}
        return
    }    

    lastResult := subjects
    for len(lastResult.Pages.NextURL) > 0 {
        fmt.Printf("Next page: %s\n", lastResult.Pages.NextURL)

        lastResult, err = getSubjectsPage(apiKey, lastResult.Pages.NextURL)
        if err != nil {
            chResult <- &Subjects{Error: err.Error()}
            return
        }

        subjects.Data = append(subjects.Data, lastResult.Data...)
    }
    
    for i := 0; i<len(subjects.Data); i++ {
        subjectsDataMap[subjects.Data[i].ID] = subjects.Data[i]
    }
    subjectsCache = subjects
    
    chResult <- subjects
}

func getSubjectsPage(apiKey string, pageUrl string) (*Subjects, error) {
    body, err := getUrl(apiKey, pageUrl)
    if err != nil {
        return nil, err
    }
    var subjects Subjects
    
    err = json.Unmarshal(body, &subjects)
    if err != nil {
        return nil, err
    }

    return &subjects, nil
}
