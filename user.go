package main

import "encoding/json"

type User struct {
    Data struct {
        CurrentVacationStartedAt string      `json:"current_vacation_started_at"`
        Level                    int64       `json:"level"`
        ProfileURL               string      `json:"profile_url"`
        StartedAt                string      `json:"started_at"`
        Subscribed               bool        `json:"subscribed"`
        Username                 string      `json:"username"`
    } `json:"data"`
    DataUpdatedAt string `json:"data_updated_at"`
    Object        string `json:"object"`
    URL           string `json:"url"`
    Error         string `json:"-"`
}

func getUser(apiKey string, chResult chan *User) {
    body, err := getUrl(apiKey, "https://wanikani.com/api/v2/user")
    if err != nil {
        chResult <- &User{Error: err.Error()}
        return
    }
    var results User
    
    err = json.Unmarshal(body, &results)
    if err != nil {
        chResult <- &User{Error: err.Error()}
        return
    }

    chResult <- &results
}
