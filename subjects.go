package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/url"

var subjectsCache *Subjects
var subjectsDataMap map[int]SubjectsData = make(map[int]SubjectsData)
var subjectsKeyMap map[string]SubjectsData = make(map[string]SubjectsData)

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

func subjectKey(subject SubjectsData) string {
    if subject.Data.Character == "" {
        return fmt.Sprintf("%s/%s", subject.Object, subject.Data.Characters)
    } else {
        return fmt.Sprintf("%s/%s", subject.Object, subject.Data.Character)
    }
}

func getSubjects(apiKey string, chResult chan *Subjects) {
    if subjectsCache != nil {
        chResult <- subjectsCache
        return
    }

    var subjects *Subjects
    raw, err := ioutil.ReadFile(GetCacheDir()+"/subjects.json")
    if (err != nil) {
        // cache miss
        subjects, err = getSubjectsPage(apiKey, "https://wanikani.com/api/v2/subjects")
        if err != nil {
            chResult <- &Subjects{Error: err.Error()}
            return
        }    
    } else {
        // cache hit
        err = json.Unmarshal(raw, &subjects)
        if err != nil {
            chResult <- &Subjects{Error: err.Error()}
            return
        }
        for _, subject := range subjects.Data {
            // fmt.Printf("Adding to map: %s\n", subjectKey(subject))
            subjectsKeyMap[subjectKey(subject)] = subject
        }

        v := url.Values{}
        v.Set("updated_after", subjects.DataUpdatedAt)
        subjects.Pages.NextURL = "https://wanikani.com/api/v2/subjects?"+v.Encode()
    }

    lastResult := subjects
    for len(lastResult.Pages.NextURL) > 0 {
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

    raw, err = json.MarshalIndent(subjects, "", "  ")
    if (err != nil) {
        fmt.Printf("Error marshalling subject cache: %s\n", err.Error())
        return
    }

    err = ioutil.WriteFile(GetCacheDir()+"/subjects.json", raw, 0644)
    if (err != nil) {
        fmt.Printf("Error writing subject cache: %s\n", err.Error())
        return
    }
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

    for _, subject := range subjects.Data {
        // fmt.Printf("Adding to map: %s\n", subjectKey(subject))
        subjectsKeyMap[subjectKey(subject)] = subject
    }

    return &subjects, nil
}
