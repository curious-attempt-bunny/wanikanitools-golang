package main

import "io/ioutil"
import "net/http"
import "time"
import "fmt"

var client *http.Client

func init() {
    http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 30
    client = &http.Client{}
}

func getUrl(apiKey string, url string) []byte {
    start := time.Now()
    req, err := http.NewRequest("GET", url, nil)
    if (err != nil) {
        panic(err)
    }
    req.Header.Add("Authorization", "Token token=" + apiKey)
    resp, err := client.Do(req)
    if (err != nil) {
        panic(err)
    }
    defer resp.Body.Close()

    if (resp.StatusCode != 200) {
        panic(fmt.Sprintf("resp.StatusCode == %d", resp.StatusCode));
    }

    body, err := ioutil.ReadAll(resp.Body)
    if (err != nil) {
        panic(err)
    }
    secs2 := time.Since(start).Seconds()
  
    fmt.Printf("%f: %s\n", secs2, url)

    return body
}