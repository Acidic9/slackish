package main

import (
	twitterOauth "github.com/mrjones/oauth"
)

var (
	twitterConsumer *twitterOauth.Consumer
	twitterTokens   = make(map[string]*twitterOauth.RequestToken)
)

func initTwitterSignIn(consumerKey, consumerSecret string) {
	twitterConsumer = twitterOauth.NewConsumer(
		consumerKey,
		consumerSecret,
		twitterOauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		},
	)
}
