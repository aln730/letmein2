package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables, one way or another
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Couldn't load .env file! Please ensure that you have a .env in this directory. Check main.go for expected variables.")
	}

	var broker, brokerMissing = os.LookupEnv("LMI_BROKER")
	var port, portMissing = os.LookupEnv("LMI_BROKER_PORT")
	var portNumber = 1883 // Set a reasonable default.
	var lmiTemplates, lmiTemplatesMissing = os.LookupEnv("LMI_TEMPLATES")
	var lmiStatic, lmiStaticMissing = os.LookupEnv("LMI_STATIC")
	var timeout, timeoutMissing = os.LookupEnv("LMI_TIMEOUT")
	var timeoutPeriod = 45 // Set a reasonable default.
	var oauthToken, oauthMissing = os.LookupEnv("LMI_OAUTH")
	var channelID, channelMissing = os.LookupEnv("LMI_CHANNEL")

	//Make sure the variables actually exist
	if !brokerMissing {
		fmt.Println("Error! MQTT Broker not specified.")
		return
	}

	if !portMissing {
		fmt.Println("Warning! MQTT Port not specified. Defaulting to 1883...")
	} else {
		portNumber, _ = strconv.Atoi(port)
	}

	if !lmiTemplatesMissing {
		fmt.Println("Error! LMI_TEMPLATES not specified.")
		return
	}

	if !lmiStaticMissing {
		fmt.Println("Error! LMI_STATIC not specified.")
		return
	}

	if !timeoutMissing {
		fmt.Println("Warning! Timeout not specified. Defaulting to ", timeoutPeriod, "...")
	} else {
		timeoutPeriod, _ = strconv.Atoi(timeout)
	}

	if !oauthMissing {
		fmt.Println("LMI_OAUTH not specified. Slack Integration is disabled! If this is not intended, please check .env!")
	}

	if !channelMissing {
		fmt.Println("LMI_CHANNEL not specified. Slack Integration is disabled! If this is not intended, please check .env!")
	}

	fmt.Println(" MQTT broker = ", broker, ", port = ", portNumber)

	// Gin Setup
	r := gin.Default()
	r.SetTrustedProxies([]string{"0.0.0.0"})

	r.LoadHTMLGlob(lmiTemplates)
	r.Static("/static", lmiStatic)

	// ===== Route definitions =====
	bot := NewSlackBot(oauthToken, channelID)
	knock := Knock{bot, 100000, broker, portNumber, timeoutPeriod}

	// Homepage
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "home.tmpl", gin.H{
			"location_map": location_map,
		})
	})

	r.GET("/knock/socket/:location", knock.handler)

	// This route sends all incoming POST requests from Slack to knock.buttonHandler
	// make sure to update the `Request Url` in the Interactivity tab in the Slack App settings, to your_server_url/actions
	// ^^^ also make sure that your server is hosted with HTTPS or Slack will be mad at you
	r.POST("/actions", buttonHandler)

	r.Run()
}
