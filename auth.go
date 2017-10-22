package main

import "os"
import "github.com/gin-gonic/gin"
import "github.com/newrelic/go-agent/_integrations/nrgin/v1"
func ApiKeyAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.DefaultQuery("api_key", os.Getenv("WANIKANI_V2_API_KEY"))
        c.Set("apiKey", apiKey)

        txn := nrgin.Transaction(c)
        if txn != nil {
            txn.SetName(c.Request.URL.Path)
            txn.AddAttribute("api_key", apiKey)
        }

        c.Next()
    }
}