package main

import (
    "log"
    "os"
    "bytes"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "math/rand"
    "sort"
    "strings"
    "time"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgin/v1"
    "github.com/gin-contrib/sessions"

    "github.com/mattes/migrate"
    _ "github.com/mattes/migrate/database/postgres"
    _ "github.com/mattes/migrate/source/file"
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

    dbMigrate();

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

            //fmt.Printf("%s, %-v\n", apiKey)

            if len(apiKey) > 0 {
                ch := make(chan *User)
                go getUser(apiKey, ch)
                user := <-ch
                if len(user.Error) > 0 {
                    renderError(c, "user", user.Error)
                    return
                }
                //fmt.Printf("%-v\n", user)

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
            withApiKey.GET("/leeches/lesson", leechesLesson)
            withApiKey.POST("/leeches/trained", postLeechesTrained)
            withApiKey.DELETE("/leeches/trained", deleteLeechesTrained)
            withApiKey.GET("/leeches", leechesList)
            withApiKey.POST("/scripts/installed", postScriptsInstalled)
            withApiKey.GET("/scripts", listScripts)
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


type Script struct {
    Author      string      `json:"author"`
    Description string      `json:"description"`
    ImgURL      interface{} `json:"img_url"`
    Installs    float64     `json:"installs"`
    Likes       float64     `json:"likes"`
    Name        string      `json:"name"`
    ScriptURL   string      `json:"script_url"`
    TopicID     float64     `json:"topic_id"`
    TopicURL    string      `json:"topic_url"`
    Version     string      `json:"version"`
    GlobalVariables []string `json:"global_variables"`
    PercentageOfUsers  float64 `json:"percentage_of_users"`
    Categories  []string    `json:"categories"`
}

type ScriptIndex struct {
    BrowserInstalls   map[string][]*Script `json:"browser_installs"`
    AvailableScripts []*Script             `json:"available_scripts"`
}

var scripts []*Script;
var nameToScript map[string]*Script;

func listScripts(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    if len(scripts) == 0 {
        raw, err := ioutil.ReadFile("static/scripts.json")
        if (err != nil) {
            c.JSON(500, gin.H{"error": err.Error()})
            return;
        }

        err = json.Unmarshal(raw, &scripts)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
        }

        nameToScript = make(map[string]*Script)
        for _, script := range scripts {
            nameToScript[script.Name] = script
        }
    }

    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer db.Close()

    rows, err := db.Query("SELECT browser_uuid, script_name, script_version, last_seen FROM scripts WHERE api_key = $1", apiKey)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    var result ScriptIndex
    result.BrowserInstalls = make(map[string][]*Script)
    result.AvailableScripts = scripts

    for rows.Next() {
        var browserUuid string;
        var scriptName string;
        var scriptVersion string;
        var lastSeen int64;
        if err := rows.Scan(&browserUuid, &scriptName, &scriptVersion, &lastSeen); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        script, present := nameToScript[scriptName]
        if !present {
            fmt.Printf("No script found with name: %s\n", scriptName)
            continue
        }

        _, present = result.BrowserInstalls[browserUuid]
        if !present {
            result.BrowserInstalls[browserUuid] = make([]*Script, 0)
        }

        result.BrowserInstalls[browserUuid] = append(result.BrowserInstalls[browserUuid], script)
    }

    var totalUsers float64
    err = db.QueryRow("SELECT COUNT(DISTINCT(api_key)) FROM scripts").Scan(&totalUsers)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    rows, err = db.Query("SELECT script_name, COUNT(DISTINCT(api_key)) as uses FROM scripts GROUP BY script_name")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    for rows.Next() {
        var scriptName string;
        var userCount float64;
        if err := rows.Scan(&scriptName, &userCount); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        script, present := nameToScript[scriptName]
        if !present {
            fmt.Printf("No script found with name: %s\n", scriptName)
            continue
        }

        script.PercentageOfUsers = 100.0 * userCount / totalUsers
    }

    c.JSON(200, result)
}

type InstalledScripts struct {
    Installed    map[string]InstalledScript `form:"installed" json:"installed"`
}

type InstalledScript struct {
    Author            string   `json:"author"`
    Description       string   `json:"description"`
    Includes          []string `json:"includes"`
    LastSeenInstalled int64  `json:"lastSeenInstalled"`
    Name              string   `json:"name"`
    Uuid              string   `json:"uuid"`
    Version           string   `json:"version"`
}

func postScriptsInstalled(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)
    browserUuid := c.Query("browser_uuid")
    var installed InstalledScripts

    err := c.BindJSON(&installed)
    if (err != nil) {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    if len(browserUuid) == 0 {
        c.JSON(500, gin.H{"error": "browserUuid query parameter required"})
        return   
    }

    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer db.Close()

    _, err = db.Exec("DELETE FROM scripts WHERE browser_uuid = $1 AND api_key = $2", browserUuid, apiKey)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    for _, script := range installed.Installed {
        _, err = db.Exec("INSERT INTO scripts (api_key, browser_uuid, script_name, script_version, last_seen) VALUES ($1, $2, $3, $4, $5)",
            apiKey, browserUuid, script.Name, script.Version, script.LastSeenInstalled)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

    }

    c.JSON(200, gin.H{"uploaded": installed})
}

func postLeechesTrained(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)
    
    var leechesTrained []LeechTrain

    err := c.BindJSON(&leechesTrained)
    if (err != nil) {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer db.Close()

    for _, leechTrain := range leechesTrained {
        _, err = db.Exec("DELETE FROM leech_train WHERE api_key = $1 AND key = $2", apiKey, leechTrain.Key)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        _, err = db.Exec("INSERT INTO leech_train (api_key, key, worst_incorrect) VALUES ($1, $2, $3)",
            apiKey, leechTrain.Key, leechTrain.WorstIncorrect)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
    }

    c.JSON(200, gin.H{"uploaded": leechesTrained})
}

func deleteLeechesTrained(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)
    
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer db.Close()

    _, err = db.Exec("DELETE FROM leech_train WHERE api_key = $1", apiKey)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"success": true})
}

func dbMigrate() {
    if os.Getenv("DATABASE_URL") == "" {
        log.Fatal("$DATABASE_URL must be set")
    }

    m, err := migrate.New(
        "file://migrations",
        os.Getenv("DATABASE_URL"))

    if err != nil {
        fmt.Printf("migrate.New failed with:\n");
        log.Fatal(err)
    }

    err = m.Up()

    if err != nil && err != migrate.ErrNoChange {
        fmt.Printf("migrate.Up failed with:\n");
        log.Fatal(err)
    }

    version, _, _ := m.Version()
    fmt.Printf("Migrations complete at version: %d\n", version)
}

func renderError(c *gin.Context, category string, error string) {
	fmt.Printf("%s.Error: %s\n", category, error)
	if strings.Contains(error, "| resp.Status = 401 Unauthorized |") {
        session := sessions.Default(c)
        session.Delete("api_key")
        session.Save()
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

func GetCacheDir() string {
    cacheDir := os.Getenv("CACHE_PATH")
    if len(cacheDir) == 0 {
        cacheDir = "data"
    }
    return cacheDir
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

    f, err := os.OpenFile(fmt.Sprintf("%s/%s_history.csv", GetCacheDir(), apiKey),
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
    
    levelToMax :=         []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    apprenticeToCount :=  []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    guruToCount :=        []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    masterToCount :=      []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    enlightenedToCount := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    burnedToCount :=      []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

    for i := 0; i<len(assignments.Data); i++ {
        assignment := assignments.Data[i].Data

        subject, isSubjectCached := subjectsDataMap[assignment.SubjectID]
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

        if assignment.SrsStage >= 1 {
            apprenticeToCount[subject.Data.Level] =  apprenticeToCount[subject.Data.Level] + 1
        }
        if assignment.SrsStage >= 5 {
            guruToCount[subject.Data.Level] =        guruToCount[subject.Data.Level] + 1
        }
        if assignment.SrsStage >= 7 {
            masterToCount[subject.Data.Level] =      masterToCount[subject.Data.Level] + 1   
        }
        if assignment.SrsStage >= 8 {
            enlightenedToCount[subject.Data.Level] = enlightenedToCount[subject.Data.Level] + 1
        }
        if assignment.SrsStage == 9 {
            burnedToCount[subject.Data.Level] =      burnedToCount[subject.Data.Level] + 1
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
        levelToMax[subject.Data.Level] = levelToMax[subject.Data.Level] + 1

        typeToProgress, exists := levelToTypeToProgress[subject.Data.Level]
        if (!exists) {
            continue
        }

        progressType := typeToProgress[subject.Object]
        progressType.Max += 1
    }

    var progress Progress

    progress.StageLevels = make(map[string]*StageLevel)
    progress.StageLevels["Apprentice"]  = buildStageLevel(apprenticeToCount, levelToMax)
    progress.StageLevels["Guru"]        = buildStageLevel(guruToCount, levelToMax)
    progress.StageLevels["Master"]      = buildStageLevel(masterToCount, levelToMax)
    progress.StageLevels["Enlightened"] = buildStageLevel(enlightenedToCount, levelToMax)
    progress.StageLevels["Burned"]      = buildStageLevel(burnedToCount, levelToMax)

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

func buildStageLevel(stageCount []int, levelMax []int) *StageLevel {
    var level int = 1
    var percentage float64 = 0
    for {
        //fmt.Printf("%d / %d : ", stageCount[level], levelMax[level])
        percentage = float64(stageCount[level])/float64(levelMax[level])
        //fmt.Printf("%g\n", percentage)
        if percentage < 0.9 {
            break
        }
        level += 1
        if level == 61 {
            percentage = 0
            break
        }
    }

    return &StageLevel{Level:level-1, PercentageNextLevel:percentage}
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

func leechesLesson(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(string)

    chStudyMaterials := make(chan *StudyMaterials)
    go getStudyMaterials(apiKey, chStudyMaterials)

    leeches, _, assignments, resourceError := getLeeches(apiKey)
    if (resourceError != nil) {
        renderError(c, resourceError.Category, resourceError.ErrorMessage)
        return
    }

    studyMaterials := <-chStudyMaterials
    if len(studyMaterials.Error) > 0 {
        renderError(c, "studyMaterials", studyMaterials.Error)
        return
    }

    studyMaterialsDataMap := make(map[int]StudyMaterialsData)
    for i := 0; i<len(studyMaterials.Data); i++ {
        studyMaterialsDataMap[studyMaterials.Data[i].Data.SubjectID] = studyMaterials.Data[i]
    }

    assignmentsDataMap := make(map[int]AssignmentsData)
    for i := 0; i<len(assignments.Data); i++ {
        assignmentsDataMap[assignments.Data[i].Data.SubjectID] = assignments.Data[i]
    }

    // exclude items

    hoursPerLevel := []int { 0, 4, 8, 24, 3*24, 7*24, 2*7*24, 30*24, 4*30*24, 0}
    leechMap := make(map[string]Leech)
    for _, leech := range leeches {
        // remove recently reviewed, or soon-by-level to be reviewed
        assignment, ok := assignmentsDataMap[leech.SubjectID]
        if (ok) {
            updatedAt, err := time.Parse(time.RFC3339, assignment.DataUpdatedAt)
            if (err == nil) {
                // nothing in the last 24 hours
                if time.Since(updatedAt).Hours() < 24 {
                    //fmt.Printf("Skipping %s since it's too recent (%d hours < 24).\n", leech.Name, int(time.Since(updatedAt).Hours()))
                    continue
                }
            }

            availableAt, err := time.Parse(time.RFC3339, assignment.Data.AvailableAt)
            if (err == nil) {
                // not too close for the srsStage
                hoursFromNow := int(availableAt.Sub(time.Now()).Hours())
                if hoursPerLevel[assignment.Data.SrsStage]/2 < hoursFromNow {
                    //fmt.Printf("Skipping %s since it's too soon to the review (stage %s, hours away %d < %d/2).\n", leech.Name, assignment.Data.SrsStageName, hoursFromNow, hoursPerLevel[assignment.Data.SrsStage])
                    continue
                }    
            }
        }

        leechMap[fmt.Sprintf("%s/%s", leech.SubjectType, leech.Name)] = leech
    }

    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer db.Close()

    rows, err := db.Query("SELECT key, worst_incorrect FROM leech_train WHERE api_key = $1", apiKey)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    for rows.Next() {
        var key string;
        var worst_incorrect int;
        if err := rows.Scan(&key, &worst_incorrect); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        leech, present := leechMap[key]
        if present {
            if leech.WorstIncorrect <= worst_incorrect {
                delete(leechMap, key)
            }
        }
    }

    leeches = make([]Leech, len(leechMap))
    var i int = 0
    for _, leech := range leechMap {
        leeches[i] = leech
        i++
    }

    result := LeechLesson{LeechesAvailable:len(leeches), LeechLessonItems:make([]LeechLessonItem, 0)}

    shuffleIndexes := rand.Perm(len(leeches))

    for i := 0; i < len(shuffleIndexes) && i < 10; i++ {
        leech := leeches[shuffleIndexes[i]]
        correctAnswers := make([]string, 0)
        tryAgainAnswers := make([]string, 0)

        subject := subjectsDataMap[leech.SubjectID]
        if leech.WorstType == "reading" {
            for _, reading := range subject.Data.Readings {
                if leech.SubjectType != "kanji" || reading.Primary {
                    correctAnswers = append(correctAnswers, reading.Reading)
                } else {
                    tryAgainAnswers = append(tryAgainAnswers, reading.Reading)
                }
            }
        } else {
            studyMaterials, ok := studyMaterialsDataMap[subject.ID]
            if ok {
                for _, synonym := range studyMaterials.Data.MeaningSynonyms {
                    correctAnswers = append(correctAnswers, synonym)
                }                
            }
            for _, meaning := range subject.Data.Meanings {
                correctAnswers = append(correctAnswers, meaning.Meaning)
            }
        }

        item := LeechLessonItem{
            Name:            leech.Name,
            SubjectType:     leech.SubjectType,
            TrainType:       leech.WorstType,
            CorrectAnswers:  correctAnswers,
            TryAgainAnswers: tryAgainAnswers}
        item.Leech.Key = fmt.Sprintf("%s/%s", leech.SubjectType, leech.Name)
        item.Leech.WorstIncorrect = leech.WorstIncorrect

        result.LeechLessonItems = append(result.LeechLessonItems, item)
        result.LeechLessonItems = append(result.LeechLessonItems, item)
        result.LeechLessonItems = append(result.LeechLessonItems, item)

        //fmt.Printf("%s\n", subjectKey(subject))

        similars := similar[subjectKey(subject)]
        shuffleIndexes2 := rand.Perm(len(similars))

        k := 0
        for j := 0; j < len(shuffleIndexes2) && k < 3; j++ {
            key := similars[shuffleIndexes2[j]]
            // fmt.Printf("  %s\n", key)
            subject, ok := subjectsKeyMap[key]
            if (!ok) {
                fmt.Printf("Nothing in map for %s\n", key)
                continue
            }
            assignment, unlocked := assignmentsDataMap[subject.ID]
            //fmt.Printf("  %s %t\n", key, unlocked)
            if unlocked && assignment.Data.SrsStage > 0 {
                correctAnswers := make([]string, 0)
                tryAgainAnswers := make([]string, 0)

                if leech.WorstType == "reading" {
                    for _, reading := range subject.Data.Readings {
                        if subject.Object != "kanji" || reading.Primary {
                            correctAnswers = append(correctAnswers, reading.Reading)
                        } else {
                            tryAgainAnswers = append(tryAgainAnswers, reading.Reading)
                        }
                    }
                } else {
                    studyMaterials, ok := studyMaterialsDataMap[subject.ID]
                    if ok {
                        for _, synonym := range studyMaterials.Data.MeaningSynonyms {
                            correctAnswers = append(correctAnswers, synonym)
                        }                
                    }                    
                    for _, meaning := range subject.Data.Meanings {
                        correctAnswers = append(correctAnswers, meaning.Meaning)
                    }
                }

                item := LeechLessonItem{
                    SubjectType:     subject.Object,
                    TrainType:       leech.WorstType,
                    CorrectAnswers:  correctAnswers,
                    TryAgainAnswers: tryAgainAnswers}
                if subject.Data.Character == "" {
                    item.Name = subject.Data.Characters
                } else {
                    item.Name = subject.Data.Character
                }
                item.Leech.Key = fmt.Sprintf("%s/%s", leech.SubjectType, leech.Name)
                item.Leech.WorstIncorrect = leech.WorstIncorrect

                result.LeechLessonItems = append(result.LeechLessonItems, item)
                k++
            }
        }
    }

    c.JSON(200, result)
}

type LeechList []Leech

func (p LeechList) Len() int { return len(p) }
func (p LeechList) Less(i, j int) bool { return p[i].WorstScore > p[j].WorstScore || (p[i].WorstScore == p[j].WorstScore && p[i].Name > p[j].Name)}
func (p LeechList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }
