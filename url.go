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

    if resp.StatusCode == 403 && strings.Contains(string(body), "Rate Limit Exceeded") {
        fmt.Printf("Rate Limit Exceeded: backing off & retrying (url %s)\n", url)
        time.Sleep(1000 * time.Millisecond)
        return getUrl(apiKey, url)
    }

    if resp.StatusCode != 200 {
        return nil, errors.New(fmt.Sprintf("apiKey = %s | url = %s | resp.StatusCode = %d | resp.Status = %s | resp.Body = %s\n", apiKey, url, resp.StatusCode, resp.Status, string(body)))
    }

    secs2 := time.Since(start).Seconds()
  
    fmt.Printf("%f: %s\n", secs2, url)

    return body, nil
}