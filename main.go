package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/acheong08/clir"
	"github.com/labstack/echo/v4"
)

var records []record

type record struct {
	Hostname string
	IP       []string
}

func main() {
	var username string
	var password string
	cli := clir.NewCli("dydns", "A dead simple dynamic DNS API", "v0.0.1")
	cli.WithFlags(
		clir.StringFlag("username", "", &username),
		clir.StringFlag("password", "", &password),
	)

	cli.Action(func() error {
		e := echo.New()

		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Get authorization
				auth := strings.Split(c.Request().Header.Get("Authorization"), " ")
				if len(auth) != 2 {
					log.Println("No auth header")
					return c.String(401, "badauth")
				}
				if auth[0] != "Basic" {
					log.Printf("Unsupported auth type: %s", auth[0])
					return c.String(401, "badauth")
				}
				// Base64 decode
				decoded, err := base64.StdEncoding.DecodeString(auth[1])
				if err != nil {
					log.Printf("Failed to decode auth: %s", err)
					return c.String(401, "badauth")
				}
				userpass := strings.Split(string(decoded), ":")
				if len(userpass) != 2 {
					log.Printf("Invalid userpass: %s", userpass)
					return c.String(401, "badauth")
				}
				if userpass[0] != username || userpass[1] != password {
					log.Printf("Username and password not matched:\nUsername:%s!=%s\nPassword:%s!=%s", userpass[0], username, userpass[1], password)
					return c.String(401, "badauth")
				}
				return next(c)
			}
		})

		e.GET("/nic/update", func(c echo.Context) error {
			hostname := c.QueryParam("hostname")
			myip := strings.Split(c.QueryParam("myip"), ",") // Optional
			if hostname == "" {
				return c.String(400, "missing hostname")
			}
			if len(myip) == 0 || myip[0] == "" {
				myip = []string{c.RealIP()}
			}
			records = append(records, record{
				Hostname: hostname,
				IP:       myip,
			})
			log.Printf("Record added %v", records[len(records)-1])
			return c.String(200, fmt.Sprintf("good %s", myip[0]))
		})

		e.GET("/nic/fetch", func(c echo.Context) error {
			limit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil || limit == 0 {
				limit = 10
			}
			if len(records) < limit {
				return c.JSON(200, records)
			}
			return c.JSON(200, records[len(records)-limit:])
		})

		return e.Start(":2005")
	})
	err := cli.Run()
	if err != nil {
		panic(err)
	}
}
