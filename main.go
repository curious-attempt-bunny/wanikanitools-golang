package main

import (
    "log"
    "os"
    "bytes"
    "fmt"
    "io"
    "net/http"
    "sort"
    "strings"
    "time"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgin/v1"
    "github.com/gin-contrib/sessions"
)

type TemplateContext struct {
    User *User
    Data interface{}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()

	if len(os.Getenv("NEW_RELIC_APP_NAME")) > 0 && len(os.Getenv("NEW_RELIC_LICENSE_KEY")) > 0 {
		config := newrelic.NewConfig(os.Getenv("NEW_RELIC_APP_NAME"), os.Getenv("NEW_RELIC_LICENSE_KEY"))
		app, err := newrelic.NewApplication(config)
		if (err != nil) {
			panic(err)
		}
		router.Use(nrgin.Middleware(app))
	}

    router.Use(CORSMiddleware())

    secret := os.Getenv("SESSION_SECRET")
    if len(secret) == 0 {
        secret = "a secret"
    }

    store := sessions.NewCookieStore([]byte(secret))
	router.Use(sessions.Sessions("session", store))

    router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

    withSessionApiKeyRedirect := router.Group("/")
    withSessionApiKeyRedirect.Use(RedirectWithSessionApiKey())
    {
    	withSessionApiKeyRedirect.GET("/", func(c *gin.Context) {
            apiKey := c.Query("api_key")

            fmt.Printf("%s, %-v\n", apiKey)

            if len(apiKey) > 0 {
                ch := make(chan *User)
                go getUser(apiKey, ch)
                user := <-ch
                if len(user.Error) > 0 {
                    renderError(c, "user", user.Error)
                    return
                }
                fmt.Printf("%-v\n", user)

    		    c.HTML(http.StatusOK, "index.tmpl.html", TemplateContext{User:user})
            } else {                
                c.HTML(http.StatusOK, "index.tmpl.html", nil)
            }
    	})

    	withApiKey := withSessionApiKeyRedirect.Group("/")
    	withApiKey.Use(ApiKeyAuth())
    	{
    		withApiKey.GET("/api/v2/subjects", apiV2Subjects)
    		withApiKey.GET("/srs/status", srsStatus)
    		withApiKey.GET("/srs/status/history.csv", srsStatusHistory)
    		withApiKey.GET("/leeches.txt", leechesTxt)
            withApiKey.GET("/leeches.json", leechesJson)
            withApiKey.GET("/level/progress", levelProgress)
            withApiKey.GET("/leeches/screensaver", leechesScreensaver)
            withApiKey.GET("/leeches", leechesList)
            withApiKey.GET("/leeches/training", leechesTraining)
    	}
    }

    router.POST("/signout", func(c *gin.Context) {
        session := sessions.Default(c)
        session.Delete("api_key")
        session.Save()
        c.Redirect(http.StatusFound, "/")
    })

	router.Run(":" + port)
}

func renderError(c *gin.Context, category string, error string) {
	fmt.Printf("%s.Error: %s\n", category, error)
	if strings.Contains(error, "| resp.Status = 401 Unauthorized |") {
		c.JSON(401, gin.H{"error": "Bad credentials"})	
	} else {
		c.JSON(500, gin.H{"error": error})
	}
}

func apiV2Subjects(c *gin.Context) {
	apiKey := c.MustGet("apiKey").(string)
	ch := make(chan *Subjects)
	go getSubjects(apiKey, ch)
	subjects := <-ch
	if len(subjects.Error) > 0 {
		renderError(c, "subjects", subjects.Error)
		return
	}

	c.JSON(200, subjects)
}

func srsStatus(c *gin.Context) {
	apiKey := c.MustGet("apiKey").(string)

    ch := make(chan *User)
    go getUser(apiKey, ch)
        
	dashboard := Dashboard{}
	dashboard.Levels.Order = []string{ "apprentice", "guru", "master", "enlightened", "burned" }

	srsLevelTotals := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    leechTotals := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    
    leeches, reviewStatistics, assignments, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
    	renderError(c, resourceError.Category, resourceError.ErrorMessage)
    	return
    }

    for i := 0; i<len(assignments.Data); i++ {
    	srsLevelTotals[assignments.Data[i].Data.SrsStage] += 1
    }

	for i := 0; i<len(leeches); i++ {
		leech := leeches[i]
		leechTotals[leech.SrsStage] += 1
	}

	sort.Sort(leeches)
	retainedLeeches := 10
	if retainedLeeches > len(leeches) {
		retainedLeeches = len(leeches)
	}

	dashboard.ReviewOrder = leeches[0:retainedLeeches]
	dashboard.LeechesTotal = len(leeches)
	dashboard.SrsLevelTotals = srsLevelTotals
    dashboard.SrsLevelLeechesTotals = leechTotals

	dashboard.Levels.Apprentice. SrsLevelTotals = srsLevelTotals[1:5]
	dashboard.Levels.Guru.       SrsLevelTotals = srsLevelTotals[5:7]
	dashboard.Levels.Master.     SrsLevelTotals = srsLevelTotals[7:8]
	dashboard.Levels.Enlightened.SrsLevelTotals = srsLevelTotals[8:9]
	dashboard.Levels.Burned.     SrsLevelTotals = srsLevelTotals[9:10]

	dashboard.Levels.Apprentice. Total = srsLevelTotals[1] + srsLevelTotals[2] + srsLevelTotals[3] + srsLevelTotals[4]
	dashboard.Levels.Guru.       Total = srsLevelTotals[5] + srsLevelTotals[6]
	dashboard.Levels.Master.     Total = srsLevelTotals[7]
	dashboard.Levels.Enlightened.Total = srsLevelTotals[8]
	dashboard.Levels.Burned.     Total = srsLevelTotals[9]

	dashboard.Levels.Apprentice. SrsLevelLeechesTotals = leechTotals[1:5]
	dashboard.Levels.Guru.       SrsLevelLeechesTotals = leechTotals[5:7]
	dashboard.Levels.Master.     SrsLevelLeechesTotals = leechTotals[7:8]
	dashboard.Levels.Enlightened.SrsLevelLeechesTotals = leechTotals[8:9]
	dashboard.Levels.Burned.     SrsLevelLeechesTotals = leechTotals[9:10]


	dashboard.Levels.Apprentice. LeechesTotal = leechTotals[1] + leechTotals[2] + leechTotals[3] + leechTotals[4]
	dashboard.Levels.Guru.       LeechesTotal = leechTotals[5] + leechTotals[6]
	dashboard.Levels.Master.     LeechesTotal = leechTotals[7]
	dashboard.Levels.Enlightened.LeechesTotal = leechTotals[8]
	dashboard.Levels.Burned.     LeechesTotal = leechTotals[9]

	c.JSON(200, dashboard)

    user := <-ch
    txn := nrgin.Transaction(c)
    
    if txn != nil {
        txn.AddAttribute("leechesTotal", dashboard.LeechesTotal)
        txn.AddAttribute("assignmentsTotal", len(assignments.Data))
        txn.AddAttribute("reviewStatisticsTotal", len(reviewStatistics.Data))
    }

    cacheDir := os.Getenv("CACHE_PATH")
    if len(cacheDir) == 0 {
    	cacheDir = "data"
    }
    f, err := os.OpenFile(fmt.Sprintf("%s/%s_history.csv", cacheDir, apiKey),
    						os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err == nil {
	    now := time.Now()
		f.Write([]byte(fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d\n",
			now.UTC().Format("2006-01-02 15:04:05"), now.Unix(), user.Data.Level,
			srsLevelTotals[1] + srsLevelTotals[2] + srsLevelTotals[3] + srsLevelTotals[4] +
			srsLevelTotals[5] + srsLevelTotals[6] + srsLevelTotals[7] + srsLevelTotals[8] + srsLevelTotals[9], dashboard.LeechesTotal,
			srsLevelTotals[1], srsLevelTotals[2], srsLevelTotals[3], srsLevelTotals[4],
			srsLevelTotals[5], srsLevelTotals[6], srsLevelTotals[7], srsLevelTotals[8], srsLevelTotals[9],
			leechTotals[1], leechTotals[2], leechTotals[3], leechTotals[4],
			leechTotals[5], leechTotals[6], leechTotals[7], leechTotals[8], leechTotals[9])))
		f.Close()
	}
}	

func srsStatusHistory(c *gin.Context) {
	apiKey := c.MustGet("apiKey").(string)

	cacheDir := os.Getenv("CACHE_PATH")
    if len(cacheDir) == 0 {
    	cacheDir = "data"
    }
    filename := fmt.Sprintf("%s/%s_history.csv", cacheDir, apiKey)
    file, err := os.Open(filename)
    if err != nil {
    	c.JSON(500, gin.H{"error":err.Error()})
    	return
    }
    defer file.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=status_history.csv"))
	c.Header("Content-Type", "text/csv")
	c.String(200, fmt.Sprintf("UTCDateTime,EpochSeconds,UserLevel,Total,LeechTotal,Apprentice1,Apprentice2,Apprentice3,"+
		"Apprentice4,Guru1,Guru2,Master,Enlightened,Burned,LeechApprentice1,LeechApprentice2,LeechApprentice3,"+
		"LeechApprentice4,LeechGuru1,LeechGuru2,LeechMaster,LeechEnlightened,LeechBurned\n"))
	io.Copy(c.Writer, file)
}

func leechesTxt(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    leeches, _, _, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
        renderError(c, resourceError.Category, resourceError.ErrorMessage)
        return
    }

    var result bytes.Buffer

    for i := 0; i < len(leeches); i++ {
        result.WriteString(fmt.Sprintf("\"%s\n(%s meaning)\";%s\n", leeches[i].Name, leeches[i].SubjectType, leeches[i].PrimaryMeaning))
        result.WriteString(fmt.Sprintf("\"%s\n(%s reading)\";%s\n", leeches[i].Name, leeches[i].SubjectType, leeches[i].PrimaryReading))
    }

    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=leeches.txt"))
    c.Header("Content-Type", "text/plain")
    c.String(200, result.String())
}

func leechesJson(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    leeches, _, _, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
        renderError(c, resourceError.Category, resourceError.ErrorMessage)
        return
    }

    c.JSON(200, leeches)
}

func levelProgress(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    chUser := make(chan *User)
    go getUser(apiKey, chUser)

    chAssignments := make(chan *Assignments)
    go getAssignments(apiKey, chAssignments)
    
    user := <-chUser
    if len(user.Error) > 0 {
        renderError(c, "user", user.Error)
        return
    }

    assignments := <-chAssignments
    if len(assignments.Error) > 0 {
        renderError(c, "assignments", assignments.Error)
        return
    } 

    levelToTypeToProgress := make(map[int]map[string]*ProgressType)
    levelToTypeToProgress[user.Data.Level - 1] = make(map[string]*ProgressType)
    levelToTypeToProgress[user.Data.Level - 1]["radical"] = &ProgressType{Level:user.Data.Level - 1, Type:"radical", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    levelToTypeToProgress[user.Data.Level - 1]["vocabulary"] = &ProgressType{Level:user.Data.Level - 1, Type:"vocabulary", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    levelToTypeToProgress[user.Data.Level - 1]["kanji"] = &ProgressType{Level:user.Data.Level - 1, Type:"kanji", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    
    levelToTypeToProgress[user.Data.Level] = make(map[string]*ProgressType)
    levelToTypeToProgress[user.Data.Level]["radical"] = &ProgressType{Level:user.Data.Level, Type:"radical", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    levelToTypeToProgress[user.Data.Level]["vocabulary"] = &ProgressType{Level:user.Data.Level, Type:"vocabulary", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    levelToTypeToProgress[user.Data.Level]["kanji"] = &ProgressType{Level:user.Data.Level, Type:"kanji", SrsLevelTotals: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}
    
    for i := 0; i<len(assignments.Data); i++ {
        assignment := assignments.Data[i].Data

        _, isSubjectCached := subjectsDataMap[assignment.SubjectID]
        if !isSubjectCached {
            fmt.Printf("Cache miss for subject ID %d - reloading\n", assignment.SubjectID)
            chSubjects := make(chan *Subjects)
            go getSubjects(apiKey, chSubjects)
            subjects := <-chSubjects
            if len(subjects.Error) > 0 {
                renderError(c, "subjects", subjects.Error)
                return
            }
        }

        typeToProgress, exists := levelToTypeToProgress[assignment.Level]
        if (!exists) {
            continue
        }

        progressType := typeToProgress[assignment.SubjectType]
        progressType.SrsLevelTotals[ assignment.SrsStage ] += 1

        if assignment.SrsStage >= 5 {
            progressType.GuruedTotal += 1
        }
    }

    for _, subject := range subjectsDataMap {
        typeToProgress, exists := levelToTypeToProgress[subject.Data.Level]
        if (!exists) {
            continue
        }

        progressType := typeToProgress[subject.Object]
        progressType.Max += 1
    }

    var progress Progress

    progress.Progresses = []*ProgressType{
        levelToTypeToProgress[user.Data.Level - 1]["radical"],
        levelToTypeToProgress[user.Data.Level - 1]["kanji"],
        levelToTypeToProgress[user.Data.Level - 1]["vocabulary"],
        levelToTypeToProgress[user.Data.Level]["radical"],
        levelToTypeToProgress[user.Data.Level]["kanji"],
        levelToTypeToProgress[user.Data.Level]["vocabulary"] }
    
    c.JSON(200, progress)

    txn := nrgin.Transaction(c)
    if txn != nil {
        txn.AddAttribute("assignmentsTotal", len(assignments.Data))
    }
}

func leechesScreensaver(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    chUser := make(chan *User)
    go getUser(apiKey, chUser)

    leeches, _, _, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
        renderError(c, resourceError.Category, resourceError.ErrorMessage)
        return
    }

    user := <-chUser
    if len(user.Error) > 0 {
        renderError(c, "user", user.Error)
        return
    }    

    c.HTML(http.StatusOK, "leeches.screensaver.tmpl.html", TemplateContext{User:user, Data:leeches})
}

func leechesList(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    chUser := make(chan *User)
    go getUser(apiKey, chUser)

    leeches, _, _, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
        renderError(c, resourceError.Category, resourceError.ErrorMessage)
        return
    }

    user := <-chUser
    if len(user.Error) > 0 {
        renderError(c, "user", user.Error)
        return
    }

    sort.Sort(leeches)
    
    c.HTML(http.StatusOK, "leeches.list.tmpl.html", TemplateContext{User:user, Data:leeches})
}

type LeechList []Leech

func (p LeechList) Len() int { return len(p) }
func (p LeechList) Less(i, j int) bool { return p[i].WorstScore > p[j].WorstScore || (p[i].WorstScore == p[j].WorstScore && p[i].Name > p[j].Name)}
func (p LeechList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
