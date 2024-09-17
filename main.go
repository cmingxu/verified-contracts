package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const u = "https://ton-source-prod-3.herokuapp.com/latestVerified"

var (
	host = flag.String("host", "localhost", "host")
	port = flag.String("port", "3306", "port")
	user = flag.String("user", "root", "user")
	pass = flag.String("pass", "password", "pass")
	db   = flag.String("db", "verified", "db")
)

type Contract struct {
	Address   string `json:"address" db:"address"`
	Compilter string `json:"compiler" db:"compiler"`
	MainFile  string `json:"mainFile" db:"main_file"`
}

func main() {
	fmt.Println("hello world")

	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *user, *pass, *host, *port, *db)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		panic(err)
	}

	for {
		proxyURL, _ := url.Parse("http://192.168.8.38:4780")
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest("GET", u, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		defer resp.Body.Close()

		contracts := []Contract{}
		err = json.NewDecoder(resp.Body).Decode(&contracts)
		if err != nil {
			continue
		}

		for _, contract := range contracts {
			rows, err := db.Query("SELECT * FROM contracts WHERE address = ?", contract.Address)
			if err != nil {
				break
			}

			if rows.Next() {
				continue
			}

			_, err = db.NamedExec("INSERT INTO contracts (address, compiler, main_file) VALUES (:address, :compiler, :main_file)", contract)
			if err != nil {
				break
			}
		}

		time.Sleep(5 * time.Minute)
	}

}
