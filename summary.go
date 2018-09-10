package main

import "encoding/json"

type Summary struct {
	Data struct {
		LessonSubjectIds []int `json:"lesson_subject_ids"`
		ReviewSubjectIds []int `json:"review_subject_ids"`
		ReviewsPerHour   []struct {
			AvailableAt string `json:"available_at"`
			SubjectIds  []int  `json:"subject_ids"`
		} `json:"reviews_per_hour"`
	} `json:"data"`
	DataUpdatedAt string `json:"data_updated_at"`
	Object        string `json:"object"`
	URL           string `json:"url"`
	Error         string `json:"-"`
}

func getSummary(apiKey string, chResult chan *Summary) {
	body, err := getUrl(apiKey, "https://api.wanikani.com/v2/summary")
	if err != nil {
		chResult <- &Summary{Error: err.Error()}
		return
	}
	var results Summary

	err = json.Unmarshal(body, &results)
	if err != nil {
		chResult <- &Summary{Error: err.Error()}
		return
	}

	chResult <- &results
}
