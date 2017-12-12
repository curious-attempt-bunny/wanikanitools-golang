package main

import "io/ioutil"
import "net/http"
import "time"
import "fmt"
import "errors"
import "strings"

var client *http.Client

type ResourceError struct {
    Category string
    ErrorMessage string
}

func init() {
    http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 30
    client = &http.Client{}
}

func getUrl(apiKey string, url string) ([]byte, error) {
    start := time.Now()
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Add("Authorization", "Token token=" + apiKey)
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode == 429 && strings.Contains(string(body), "Rate Limit Exceeded") {
        fmt.Printf("Rate Limit Exceeded failing: %s\n", url)
        return nil, errors.New("Rate Limit Exceeded")
    }

    if resp.StatusCode != 200 {
        return nil, errors.New(fmt.Sprintf("apiKey = %s | url = %s | resp.StatusCode = %d | resp.Status = %s | resp.Body = %s\n", apiKey, url, resp.StatusCode, resp.Status, string(body)))
    }

    secs2 := time.Since(start).Seconds()
  
    fmt.Printf("%f: %s (%s)\n", secs2, url, apiKey)

    return body, nil
}