package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	ConfigDir = ".config/gmailcheck"
)

func getClient(config *oauth2.Config) *http.Client {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	tokFile := path.Join(homedir, ConfigDir, "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = tokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authUrl := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Please go to the following URL, then type the authorisation code: \n%v\n ", authUrl)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorisation string: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from the web: %v", err)
	}

	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, tok *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Can't cache token: %v\n", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tok)
}

func main() {
	b, err := ioutil.ReadFile(path.Join(ConfigDir, "credentials.json"))
	if err != nil {
		log.Fatalf("Couldn't read credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Couldn't parse configuration: %v", err)
	}

	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Couldn't get gmail from client: %v", err)
	}

	user := "me"

	r, err := srv.Users.Messages.List(user).Q("label:inbox is:unread").Do()
	if err != nil {
		log.Fatalf("Couldn't retrieve labels: %v", err)
	}

	unreadCount := len(r.Messages)

	json.NewEncoder(os.Stdout).Encode(struct {
		Text    string `json:"text,omitempty"`
		Alt     string `json:"alt,omitempty"`
		Tooltip string `json:"tooltip,omitempty"`
	}{
		fmt.Sprintf("%d ✉️", unreadCount),
		"mail",
		fmt.Sprintf("%d unread mail", unreadCount),
	})
}
