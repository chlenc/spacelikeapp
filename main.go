// main.go

package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type App struct {
	db *gorm.DB
}
type user struct {
	gorm.Model
	Username  string `json:"username"`
	Password  string `json:"-"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

type likeEvent struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"-"`
	Target   string `json:"target"`
	Error    string `json:"error"`
}

func main() {
	//db, err := gorm.Open("postgres", "host=localhost port=5432 dbname=postgres user=alesenka sslmode=disable")
	url := os.Getenv("DATABASE_URL")
	connection, _ := pq.ParseURL(url)
	connection += " sslmode=require"

	db, err := gorm.Open("postgres", connection)

	if err != nil {
		log.Println("failed to connect database", err)
		return
	}
	defer db.Close()

	db.AutoMigrate(&user{}, &likeEvent{})

	var app = App{db}

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gin.Logger())

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")

	app.initializeRoutes(r)

	r.Run()
}

//routes.go

func (app *App) initializeRoutes(r *gin.Engine) {
	r.Use(setUserStatus())
	r.GET("/", func(c *gin.Context) {
		if !loggedInCheck(c) {
			render(c, gin.H{"IsLogin": false}, "login.html")
		}
	})
	r.GET("/login", func(c *gin.Context) {
		if !loggedInCheck(c) {
			render(c, gin.H{"IsLogin": false}, "login.html")
		}
	})
	r.POST("/login", app.performLogin)

	r.GET("/registration", func(c *gin.Context) {
		if !loggedInCheck(c) {
			render(c, gin.H{"IsLogin": false}, "registration.html")
		}
	})
	r.POST("/registration", app.register)

	r.GET("/likes", func(c *gin.Context) {
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)
		render(c, gin.H{"IsLogin": loggedIn}, "likes.html")
	})
	r.POST("/likes", app.doLike)

	r.GET("/logout", logout)

	r.POST("/checkUsername", app.checkUsername)

	//====
	r.NoRoute(func(c *gin.Context) {
		render(c, gin.H{
			"Url": c.Request.URL.Path,
		}, "404.html")
	})

}

func loggedInCheck(c *gin.Context) bool {
	loggedInInterface, _ := c.Get("is_logged_in")
	loggedIn := loggedInInterface.(bool)
	if loggedIn {
		render(c, gin.H{
			"User":    "anonymous",
			"IsError": false,
			"IsLogin": loggedIn,
		}, "profile.html")
		return true
	}

	return false
}

func render(c *gin.Context, data gin.H, templateName string) {
	//loggedInInterface, _ := c.Get("is_logged_in")
	//data["is_logged_in"] = loggedInInterface.(bool)

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		//Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		//Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		//Respond with HTML
		c.HTML(http.StatusOK, templateName, data)

	}
}

func logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "", "", false, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (app *App) performLogin(c *gin.Context) {
	username := c.PostForm("Username")
	password := c.PostForm("Password")

	var temp user
	app.db.First(&temp, "username = ?", username)
	if (temp != user{}) && (temp.Password == password) {

		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)
		c.Set("is_logged_in", true)

		render(c, gin.H{"IsLogin": true}, "profile.html")

	} else {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"IsError": true})
	}
}
func (app *App) register(c *gin.Context) {
	var newUser = user{
		Username:  c.PostForm("Username"),
		Password:  c.PostForm("Password1"),
		FirstName: c.PostForm("FirstName"),
		LastName:  c.PostForm("LastName"),
		Email:     c.PostForm("Email"),
	}
	if _, err := app.registerNewUser(newUser); err == nil {
		token := generateSessionToken()
		c.SetCookie("token", token, 3600, "", "", false, true)
		c.Set("is_logged_in", true)

		render(c, gin.H{"IsLogin": true}, "profile.html")

	} else {
		c.HTML(http.StatusBadRequest, "registration.html", gin.H{
			"IsError":      true,
			"ErrorMessage": err.Error()})

	}
}

func (app *App) registerNewUser(newUser user) (*user, error) {

	var temp user
	app.db.First(&temp, "username = ?", newUser.Username)

	if strings.TrimSpace(newUser.Password) == "" {
		return nil, errors.New("The password can't be empty")
	} else if (temp != user{}) {
		return nil, errors.New("The username isn't available")
	}
	app.db.Create(&newUser)

	return &temp, nil
}

func (app *App) checkUsername(c *gin.Context) {
	username := c.PostForm("Username")
	var temp user
	var answer = false
	app.db.First(&temp, "username = ?", username)
	if (temp == user{}) {
		answer = true
	}

	c.JSON(200, gin.H{
		"success": answer,
		"IsError": false,
	})
}

func setUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie("token"); err == nil || token != "" {
			c.Set("is_logged_in", true)
		} else {
			c.Set("is_logged_in", false)
		}
	}
}
func (app *App) doLike(c *gin.Context) {
	var event = likeEvent{
		Username: c.PostForm("login"),
		Password: c.PostForm("password"),
		Target:   c.PostForm("target"),
		Error:    "",
	}
	IsError, err := likeTarget(c)
	c.JSON(200, gin.H{
		"error":   err,
		"IsError": IsError,
	})
	event.Error = err
	app.db.Create(&event)
}

func generateSessionToken() string {
	return strconv.FormatInt(rand.Int63(), 16)
}
