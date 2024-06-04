package main

import (
	"encoding/base64"
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
					return c.String(401, "Unauthorized")
				}
				if auth[0] != "Basic" {
					return c.String(401, "Unsupported authorization type")
				}
				// Base64 decode
				decoded, err := base64.StdEncoding.DecodeString(auth[1])
				if err != nil {
					return c.String(401, "Invalid authorization")
				}
				userpass := strings.Split(string(decoded), ":")
				if len(userpass) != 2 {
					return c.String(401, "Missing username or password")
				}
				if userpass[0] != username || userpass[1] != password {
					return c.String(401, "Invalid username or password")
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
			return c.String(200, "")
		})

		e.GET("/nic/fetch", func(c echo.Context) error {
			limit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil || limit == 0 {
				limit = 10
			}
			return c.JSON(200, records[len(records)-limit:])
		})

		return nil
	})
	cli.Run()
}
