package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"github.com/karan/TakeoverBot/tweettracker"
)

const (
	SitesCSVURL                 = "https://github.com/MassMove/AttackVectors/raw/master/LocalJournals/sites.csv"
	LocalSitesFilePath          = "/tmp/sites.csv"
	SearchSleep                 = 10 * time.Second
	TweetSleep                  = 60 * time.Second
	TweetMessageWithoutLocation = "@%s That website is part of a multi-billion dollar propaganda effort to divide and manipulate people. How did you come across it? Want me to shut up? Just yell."
	TweetMessageWithLocation    = "@%s That website is part of a multi-billion dollar propaganda effort to divide and manipulate people in %s. How did you come across it? Want me to shut up? Just yell."
)

type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

type CSVLine struct {
	awsOrigin               string
	domain                  string
	state                   string
	lat                     string
	lng                     string
	locationVerified        string
	httpResponseCode        string
	contentLength           string
	facebookUrl             string
	siteName                string
	twitterUsername         string
	itunesAppStoreUrl       string
	twitterAccountCreatedAt string
	twitterUserId           string
	twitterFollowers        string
	twitterFollowing        string
	twitterTweets           string
	siteOperator            string
}

// getClient is a helper function that will return a twitter client
// that we can subsequently use to send tweets, or to stream new tweets
// this will take in a pointer to a Credential struct which will contain
// everything needed to authenticate and return a pointer to a twitter Client
// or an error
func getClient(creds *Credentials) (*twitter.Client, error) {
	// Pass in your consumer key (API Key) and your Consumer Secret (API Secret)
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	// Pass in your Access Token and your Access Token Secret
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	// we can retrieve the user and verify if the credentials
	// we have used successfully allow us to log in!
	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}

	log.Printf("User's ACCOUNT:\n%+v\n", user)
	return client, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	log.Printf("Downloading %s to %s\n", url, filepath)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func ParseData(filepath string) ([]*CSVLine, error) {
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	lines, _ := reader.ReadAll()

	log.Printf("Read CSV, and found %d lines\n", len(lines))

	lines = append(lines[:0], lines[0+1:]...)

	parsedData := []*CSVLine{}
	for _, line := range lines {
		parsedData = append(parsedData, &CSVLine{
			awsOrigin:               line[0],
			domain:                  line[1],
			state:                   line[2],
			lat:                     line[3],
			lng:                     line[4],
			locationVerified:        line[5],
			httpResponseCode:        line[6],
			contentLength:           line[7],
			facebookUrl:             line[8],
			siteName:                line[9],
			twitterUsername:         line[10],
			itunesAppStoreUrl:       line[11],
			twitterAccountCreatedAt: line[12],
			twitterUserId:           line[12],
			twitterFollowers:        line[14],
			twitterFollowing:        line[15],
			twitterTweets:           line[16],
			siteOperator:            line[17],
		})
	}
	return parsedData, nil
}

func main() {
	tweettracker.Init()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	searchCreds := Credentials{
		AccessToken:       os.Getenv("SEARCH_ACCESSTOKEN"),
		AccessTokenSecret: os.Getenv("SEARCH_ACCESSTOKENSECRET"),
		ConsumerKey:       os.Getenv("SEARCH_CONSUMERKEY"),
		ConsumerSecret:    os.Getenv("SEARCH_CONSUMERSECRET"),
	}

	tweetCreds := Credentials{
		AccessToken:       os.Getenv("TWEET_ACCESSTOKEN"),
		AccessTokenSecret: os.Getenv("TWEET_ACCESSTOKENSECRET"),
		ConsumerKey:       os.Getenv("TWEET_CONSUMERKEY"),
		ConsumerSecret:    os.Getenv("TWEET_CONSUMERSECRET"),
	}

	searchClient, err := getClient(&searchCreds)
	if err != nil {
		log.Fatalf("Error getting Twitter search client: %+v\n", err)
	}
	tweetClient, err := getClient(&tweetCreds)
	if err != nil {
		log.Fatalf("Error getting Twitter tweet client: %+v\n", err)
	}

	if err := DownloadFile(LocalSitesFilePath, SitesCSVURL); err != nil {
		log.Fatalf("Error downloading sites file: %+v\n", err)
	}

	parsedData, err := ParseData(LocalSitesFilePath)
	if err != nil {
		log.Fatalf("Error parsing data: %+v", err)
	}
	log.Println("shuffling data")
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(parsedData), func(i, j int) { parsedData[i], parsedData[j] = parsedData[j], parsedData[i] })

	for _, dataLine := range parsedData {
		// Search for each domain with max number of count.
		search, _, err := searchClient.Search.Tweets(&twitter.SearchTweetParams{
			Query: fmt.Sprintf("%s", dataLine.domain),
			Count: 100,
		})
		if err != nil {
			log.Fatalf("Error while searching for %s: %+v", dataLine.domain, err)
		}

		// Reply to tweets
		if len(search.Statuses) > 0 {
			log.Printf("Found %d tweets for domain %s\n", len(search.Statuses), dataLine.domain)
			log.Printf("Full record: %+v\n", dataLine)

			for _, tweet := range search.Statuses {
				// TODO: remove old tweets
				// Remove retweets
				if tweet.RetweetedStatus != nil {
					log.Printf("Skipping https://twitter.com/%s/status/%s since it's a retweet", tweet.User.ScreenName, tweet.IDStr)
					continue
				}
				// Check if id exists in data
				log.Printf("Checking if already replied to https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr)
				if tweettracker.Exists(tweet.IDStr) {
					continue
				}

				replyText := fmt.Sprintf(TweetMessageWithoutLocation, tweet.User.ScreenName)
				if dataLine.state != "" {
					replyText = fmt.Sprintf(TweetMessageWithLocation, tweet.User.ScreenName, dataLine.state)
				}

				replyTweet, _, err := tweetClient.Statuses.Update(replyText, &twitter.StatusUpdateParams{
					InReplyToStatusID: tweet.ID,
				})
				if err != nil {
					log.Println(err)
				}
				log.Printf("Successfully tweeted https://twitter.com/%s/status/%s", replyTweet.User.ScreenName, replyTweet.IDStr)

				tweettracker.Add(&tweettracker.DataLine{
					UserName:      tweet.User.ScreenName,
					Domain:        dataLine.domain,
					ReplyTweetId:  replyTweet.IDStr,
					IDStr:         tweet.IDStr,
					CreatedAt:     tweet.CreatedAt,
					FavoriteCount: strconv.Itoa(tweet.FavoriteCount),
					ReplyCount:    strconv.Itoa(tweet.ReplyCount),
					RetweetCount:  strconv.Itoa(tweet.RetweetCount),
					QuoteCount:    strconv.Itoa(tweet.QuoteCount),
					FullText:      tweet.FullText,
				})

				log.Printf("Sleeping for %s...\n", TweetSleep.String())
				time.Sleep(TweetSleep)
				log.Printf("Moving on to a different domain...")
				break
			}
		}

		log.Printf("Sleeping for %s...\n", SearchSleep.String())
		time.Sleep(SearchSleep)
	}
}
