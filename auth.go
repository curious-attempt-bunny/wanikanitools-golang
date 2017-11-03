package main

import "net/http"
import "os"
import "github.com/gin-gonic/gin"
import "github.com/newrelic/go-agent/_integrations/nrgin/v1"
import "github.com/gin-contrib/sessions"

func RedirectWithSessionApiKey() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.Query("api_key")

        session := sessions.Default(c)

        var redirect = false

        if len(apiKey) > 0 {
            session.Set("api_key", apiKey)
            session.Save()
        } else {
            value := session.Get("api_key")
            if (value != nil) {
                apiKey = value.(string)
                redirect = true
            }
        }

        if (redirect) {
            parameters := c.Request.URL.Query()
            parameters.Add("api_key", apiKey)
            c.Request.URL.RawQuery = parameters.Encode()
            c.Redirect(http.StatusFound, c.Request.URL.String())
            c.Abort()
            return
        }

        c.Next()
    }
}

func ApiKeyAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.DefaultQuery("api_key", os.Getenv("WANIKANI_V2_API_KEY"))
        c.Set("apiKey", apiKey)

        txn := nrgin.Transaction(c)
        if txn != nil {
            txn.SetName(c.Request.URL.Path)
            txn.AddAttribute("api_key", apiKey)
        }

        ch := make(chan *User)
        go getUser(apiKey, ch)

        c.Next()
        
        user := <-ch
        if txn != nil {
            txn.AddAttribute("user_username", user.Data.Username)
            txn.AddAttribute("user_level", user.Data.Level)
        }
    }
}