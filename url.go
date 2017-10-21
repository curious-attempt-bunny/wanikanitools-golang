package main

import "io/ioutil"
import "net/http"
import "os"
import "time"
import "fmt"

func getUrl(url string) []byte {
    start := time.Now()
    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("Authorization", "Token token=" + os.Getenv("WANIKANI_V2_API_KEY"))
    resp, _ := client.Do(req)
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    secs2 := time.Since(start).Seconds()
  
    fmt.Printf("%f: %s\n", secs2, url)

    return body
}