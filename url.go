package main

import "io/ioutil"
import "net/http"
import "time"
import "fmt"

var client *http.Client

func init() {
    client = &http.Client{}
}

func getUrl(apiKey string, url string) []byte {
    start := time.Now()
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("Authorization", "Token token=" + apiKey)
    resp, _ := client.Do(req)
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    secs2 := time.Since(start).Seconds()
  
    fmt.Printf("%f: %s\n", secs2, url)

    return body
}