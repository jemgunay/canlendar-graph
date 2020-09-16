package calendar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

var (
	CalendarName = "Units Consumed"
	credsFile    = "config/credentials.json"
)

func Fetch() ([]*calendar.Event, error) {
	b, err := ioutil.ReadFile(credsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %s", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %s", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err)
	}

	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %s", err)
	}

	// get a list of calendars
	list, err := srv.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve calendar list: %s", err)
	}

	// iterate over all calendars and locate the corresponding ID for the target calendar name
	var calendarID string
	for _, item := range list.Items {
		if CalendarName == item.Summary {
			calendarID = item.Id
			break
		}
	}

	// validate that calendar ID was found for target calendar
	if calendarID == "" {
		return nil, fmt.Errorf("failed to find ID for the %s calendar", CalendarName)
	}

	// request all events for target calendar
	req := srv.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		OrderBy("startTime")

	events, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %s", err)
	}

	if len(events.Items) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	return events.Items, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%s\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %s", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %s", err)
	}
	return token, nil
}

// Retrieves a token from a local file.
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

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
