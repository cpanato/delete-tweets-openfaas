package function

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Handle a serverless request
func Handle(req []byte) bool {

	apiKey, err := getAPISecret("TWITTER_API_KEY")
	if err != nil || apiKey == "" {
		log.Fatal("Twitter API key required")
		return false
	}

	apiSecret, err := getAPISecret("twitter_api_secret")
	if err != nil || apiSecret == "" {
		log.Fatal("Twitter API secret required")
		return false
	}

	accessToken, err := getAPISecret("twitter_access_token")
	if err != nil || accessToken == "" {
		log.Fatal("Twitter access token required")
		return false
	}

	accessTokenSecret, err := getAPISecret("twitter_access_token_secret")
	if err != nil || accessTokenSecret == "" {
		log.Fatal("Twitter access token secret required")
		return false
	}

	twitterUserName, err := getAPISecret("twitter_username")
	if err != nil || twitterUserName == "" {
		log.Fatal("Twitter username required")
		return false
	}

	tweetsToIgnore := strings.Split(os.Getenv("TWEETS_IGNORE"), ",")

	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	timeToDelete := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	fmt.Printf("will delete tweets older then %+v\n", timeToDelete)
	query := fmt.Sprintf("from:%s", twitterUserName)

	// search tweets
	searchTweetParams := &twitter.SearchTweetParams{
		Query: query,
		Until: timeToDelete,
		Count: 500,
	}

	search, resp, err := client.Search.Tweets(searchTweetParams)
	fmt.Printf("TOTAL TWEETS:\n%+v\n", len(search.Statuses))
	fmt.Printf("RESP:\n%+v\n", resp)
	fmt.Printf("Err:\n%+v\n", err)
	fmt.Printf("***************\n")

	for _, status := range search.Statuses {
		flag := false
		for _, tweet := range tweetsToIgnore {
			if status.IDStr == tweet {
				fmt.Printf("tweet is in the whitelist - %v\n", status.ID)
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		fmt.Printf("Will delete: %+v - %+v\n", status.ID, status)
		// status destroy
		params := &twitter.StatusDestroyParams{TrimUser: twitter.Bool(false)}
		tweet, resp, err := client.Statuses.Destroy(status.ID, params)
		fmt.Printf("STATUSES DESTROY:\n%+v\n", tweet)
		fmt.Printf("RESP:\n%+v\n", resp)
		fmt.Printf("Err:\n%+v\n", err)
	}

	search, _, _ = client.Search.Tweets(searchTweetParams)
	fmt.Printf("TOTAL TWEETS AFTER DELETION:\n%+v\n", len(search.Statuses))

	return true
}

func getAPISecret(secretName string) (string, error) {
	var secretBytes []byte
	var err error
	secretBytes, err = ioutil.ReadFile("/var/openfaas/secrets/" + secretName)
	if err != nil {
		// read from the original location for backwards compatibility with openfaas <= 0.8.2
		secretBytes, err = ioutil.ReadFile("/run/secrets/" + secretName)
	}

	return string(secretBytes), err
}
