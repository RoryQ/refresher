package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var GcpLogin = os.Getenv("GCP_LOGIN")
var RefreshDebug = os.Getenv("REFRESH_DEBUG")

func main() {
	refresh(GcpLogin)
}

func loginCachePath() string {
	usr, _ := user.Current()
	return path.Join(usr.HomeDir, ".cache/refresher/cache")
}

func refresh(login string) {
	expiry, found := getGcpExpiryFromCache(login)

	if !found {
		printDebug("login "+login+" not found in cache")
		expiry = getGcpExpiryFromGcloud(login)
		saveGcpExpiryToCache(login, expiry)
	}

	printDebug("expiry: "+expiry.String())
	printDebug("now UTC: "+time.Now().UTC().String())

	if time.Now().UTC().After(expiry) {
		gcloudLogin()
		expiry = getGcpExpiryFromGcloud(login)
		saveGcpExpiryToCache(login, expiry)
	} else {
		printDebug("nothing to do")
	}
}

func printDebug(msg string) {
	if RefreshDebug == "1" {
		fmt.Println(msg)
	}
}


func gcloudLogin() {
	gcloud := exec.Command("gcloud", "auth", "login", "--update-adc")
	var out bytes.Buffer
	gcloud.Stdout = &out
	gcloud.Run()
}

func getGcpExpiryFromGcloud(login string) time.Time {
	gcloud := exec.Command("gcloud", "auth", "describe", login)
	var out bytes.Buffer
	gcloud.Stdout = &out
	gcloud.Run()

	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, "token_expiry") {
			dateString := strings.Trim(strings.ReplaceAll(line, "token_expiry: ", ""), "'")
			date, err := time.Parse(time.RFC3339, dateString)
			if err != nil {
				log.Fatal(err.Error())
			}
			return date
		}
	}

	log.Fatal("no expiry found")
	return time.Time{}
}

type item struct {
	Expiry time.Time
	Created time.Time
}

type cache struct {
	GCP map[string]item
}

func getGcpExpiryFromCache(login string) (expiry time.Time, found bool) {
	bytes, err := os.ReadFile(loginCachePath())
	if err != nil {
		printDebug(err.Error())
		return time.Time{}, false
	}
	c := cache{}
	json.Unmarshal(bytes, &c)
	k, found := c.GCP[login]
	return k.Expiry, true
}

func saveGcpExpiryToCache(login string, expiry time.Time) {
	bytes, err := os.ReadFile(loginCachePath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err.Error())
	}
	c := cache{}
	json.Unmarshal(bytes, &c)
	if c.GCP == nil {
		c.GCP = make(map[string]item)
	}
	c.GCP[login] = item {
		Expiry: expiry,
		Created: time.Now().UTC(),
	}
	bytes, err = json.Marshal(c)

	err = os.MkdirAll(filepath.Dir(loginCachePath()), 0700)
	if err != nil {
		printDebug(err.Error())
	}
	err = os.WriteFile(loginCachePath(), bytes, 0644)
	if err != nil {
		printDebug(err.Error())
	}
}
