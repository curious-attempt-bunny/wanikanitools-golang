package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/url"

type StudyMaterials struct {
	Data          []StudyMaterialsData `json:"data"`
	DataUpdatedAt string               `json:"data_updated_at"`
	Object        string               `json:"object"`
	Pages         struct {
		NextURL     string `json:"next_url"`
		PerPage     int    `json:"per_page"`
		PreviousURL string `json:"previous_url"`
	} `json:"pages"`
	TotalCount int    `json:"total_count"`
	URL        string `json:"url"`
	Error      string `json:"-"`
}

type StudyMaterialsData struct {
	Data struct {
		CreatedAt       string   `json:"created_at"`
		MeaningNote     string   `json:"meaning_note"`
		MeaningSynonyms []string `json:"meaning_synonyms"`
		ReadingNote     string   `json:"reading_note"`
		SubjectID       int      `json:"subject_id"`
		SubjectType     string   `json:"subject_type"`
	} `json:"data"`
	DataUpdatedAt string `json:"data_updated_at"`
	ID            int    `json:"id"`
	Object        string `json:"object"`
	URL           string `json:"url"`
}

func getStudyMaterials(apiKey string, chResult chan *StudyMaterials) {
	var results *StudyMaterials

	apiEntity := "study_materials"
	uri := "https://api.wanikani.com/v2/" + apiEntity
	cacheFile := fmt.Sprintf("%s/%s_%s.json", GetCacheDir(), apiKey, apiEntity)
	raw, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		// cache miss
		results, err = getStudyMaterialsPage(apiKey, uri)
		if err != nil {
			chResult <- &StudyMaterials{Error: err.Error()}
			return
		}
	} else {
		// cache hit
		err = json.Unmarshal(raw, &results)
		if err != nil {
			chResult <- &StudyMaterials{Error: err.Error()}
			return
		}
		v := url.Values{}
		v.Set("updated_after", results.DataUpdatedAt)
		results.Pages.NextURL = uri + "?" + v.Encode()
	}

	itemDataMap := make(map[int]StudyMaterialsData)
	for _, item := range results.Data {
		itemDataMap[item.Data.SubjectID] = item
	}

	lastResult := results
	for len(lastResult.Pages.NextURL) > 0 {
		lastResult, err = getStudyMaterialsPage(apiKey, lastResult.Pages.NextURL)
		if err != nil {
			chResult <- &StudyMaterials{Error: err.Error()}
			return
		}

		for _, item := range lastResult.Data {
			itemDataMap[item.Data.SubjectID] = item
		}
	}

	results.Data = make([]StudyMaterialsData, 0, len(itemDataMap))
	for _, item := range itemDataMap {
		results.Data = append(results.Data, item)
	}

	chResult <- results

	raw, err = json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling %s cache: %s\n", apiEntity, err.Error())
		return
	}

	err = ioutil.WriteFile(cacheFile, raw, 0644)
	if err != nil {
		fmt.Printf("Error writing %s cache: %s\n", apiEntity, err.Error())
		return
	}
}

func getStudyMaterialsPage(apiKey string, pageUrl string) (*StudyMaterials, error) {
	body, err := getUrl(apiKey, pageUrl)
	if err != nil {
		return nil, err
	}

	var results StudyMaterials

	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
