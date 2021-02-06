package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"os"

	gorm "gorm.io/gorm"
)

type mXConfig struct {
	BaseURL  string
	ClientID string
	APIKey   string
	Headers  [3][2]string
	Client   *http.Client
	DB       *gorm.DB
}

var mxConfig = mXConfig{
	BaseURL:  "https://int-api.mx.com/",
	ClientID: "579aeb7a-54fa-4245-92ed-a13912ca622a",
	APIKey:   "3c4fb671fd9806485aef1111a75592dd28efee4d",
}

func initHeaders() {
	sEnc := b64.StdEncoding.EncodeToString([]byte(mxConfig.ClientID + ":" + mxConfig.APIKey))
	mxConfig.Headers = [3][2]string{
		{"Accept", `application/vnd.mx.api.v1+json`},
		{"Content-Type", `application/json`},
		{"Authorization", "Basic " + sEnc},
	}
}

//User ...
type User struct {
	gorm.Model
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	MXID     string `json:"mxid"`
}

//MXUser ...
type MXUser struct {
	User struct {
		Email      string `json:"email"`
		GUID       string `json:"guid"`
		ID         string `json:"id"`
		IsDisabled string `json:"is_disabled"`
		Metadata   string `json:"metadata"`
	} `json:"user"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
		//log.Fatal("$PORT must be set")
	}
	mxConfig.DB = initDB()

	router := gin.New()
	router.Use(gin.Logger())

	mxConfig.Client = &http.Client{}
	initHeaders()
	router.GET("/users", handleUsers)
	router.GET("/yourcompanyusers", handleYourCompanyUsers)
	router.POST("/createyourcompanyuser", handleCreateYourCompanyUser)
	router.POST("/createmxuser/:device_id", handleCreateMXUser)
	router.GET("/getconnectwidget/:device_id", handleGetConnectWidget)
	router.GET("/getaccounts/:device_id", handleGetAccounts)
	router.Run(":" + port)
}

//createMYCompanyUser
//getMyCompanyUsers
//get top institutions
//search for my institution
//if first insitution to add then create MX user
//create new member
//gets first credentials
//mfa
//error handling
//show all accounts
//show transactions
//show user membership in all institutions

func users(client *http.Client) string {
	req, err := http.NewRequest("GET", mxConfig.BaseURL+"users", nil)
	for _, header := range mxConfig.Headers {
		req.Header.Add(header[0], header[1])
	}
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Sprintf("Error: %v\n", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return fmt.Sprintf("%s", body)

}
func handleUsers(c *gin.Context) {
	c.String(http.StatusOK, users(mxConfig.Client))
}
func handleCreateMXUser(c *gin.Context) {

	id := c.Params.ByName("device_id")

	var user User
	mxConfig.DB.Where("device_id = ?", id).First(&user)
	if user.DeviceID == id && len(user.MXID) == 0 {
		body := "{" +
			"\"user\": {" +
			"\"id\": \"" + id + "\"," +
			"\"is_disabled\": false," +
			"\"email\": \"totally.fake.email@notreal.com\"," +
			"\"metadata\": \"" + user.Name + "\"" +
			"}" +
			"}"
		httpReq, _ := http.NewRequest("POST", mxConfig.BaseURL+"users", strings.NewReader(body))
		for _, header := range mxConfig.Headers {
			httpReq.Header.Add(header[0], header[1])
		}
		resp, err := mxConfig.Client.Do(httpReq)
		var p MXUser
		err = json.NewDecoder(resp.Body).Decode(&p)
		user.MXID = p.User.GUID
		log.Println(err)
		mxConfig.DB.Save(&user)
	}
	c.JSON(http.StatusOK, user)
}
func handleGetConnectWidget(c *gin.Context) {
	id := c.Params.ByName("device_id")
	var user User
	mxConfig.DB.Where("device_id = ?", id).First(&user)

	if user.DeviceID == id && len(user.MXID) != 0 {
		body := "{" +
			"\"widget_url\": {" +
			"\"widget_type\": \"connect_widget\"," +
			"\"color_scheme\": \"dark\"," +
			"\"is_mobile_webview\": true" +
			"}" +
			"}"
		httpReq, _ := http.NewRequest("POST", mxConfig.BaseURL+"users/"+user.MXID+"/widget_urls", strings.NewReader(body))
		for _, header := range mxConfig.Headers {
			httpReq.Header.Add(header[0], header[1])
		}
		resp, err := mxConfig.Client.Do(httpReq)
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		c.String(http.StatusOK, bodyString)
		return
	}
	c.String(http.StatusOK, "{\"error\":\"error\"}")
}
func handleGetAccounts(c *gin.Context) {
	id := c.Params.ByName("device_id")
	var user User
	mxConfig.DB.Where("device_id = ?", id).First(&user)

	if user.DeviceID == id && len(user.MXID) != 0 {
		httpReq, _ := http.NewRequest("GET", mxConfig.BaseURL+"users/"+user.MXID+"/accounts", nil)
		for _, header := range mxConfig.Headers {
			httpReq.Header.Add(header[0], header[1])
		}
		resp, err := mxConfig.Client.Do(httpReq)
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		c.String(http.StatusOK, bodyString)
		return
	}
	c.String(http.StatusOK, "{\"error\":\"error\"}")
}
func handleYourCompanyUsers(c *gin.Context) {
	var users []User
	mxConfig.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func handleCreateYourCompanyUser(c *gin.Context) {
	var p User
	c.BindJSON(&p)
	var user User
	mxConfig.DB.Where("device_id = ?", p.DeviceID).First(&user)
	if len(user.DeviceID) == 0 {
		// Create a new user in our database.
		mxConfig.DB.Create(&p)
	} else {
		p = user
	}

	c.JSON(http.StatusOK, p)
}
