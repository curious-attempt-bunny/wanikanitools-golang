package main

type Subjects struct {
    Data []struct {
        Data struct {
            Character       string `json:"character"`
            CharacterImages []struct {
                ContentType string `json:"content_type"`
                URL         string `json:"url"`
            } `json:"character_images"`
            Characters          string  `json:"characters"`
            ComponentSubjectIds []int64 `json:"component_subject_ids"`
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
        ID            int64  `json:"id"`
        Object        string `json:"object"`
        URL           string `json:"url"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    Pages         struct {
        Current     int64       `json:"current"`
        Last        int64       `json:"last"`
        NextURL     string      `json:"next_url"`
        PreviousURL interface{} `json:"previous_url"`
    } `json:"pages"`
    TotalCount int64  `json:"total_count"`
    URL        string `json:"url"`
}