package main

const (
	accountTypeEmail   = "email"
	accountTypeGoogle  = "google"
	accountTypeTwitter = "twitter"
	accountTypeGithub  = "github"
)

type user struct {
	ID                   int
	AccountID            string
	AccountType          string
	Email                string
	Password             string
	DisplayName          string
	DisplayNameGenerated bool
	FirstName            string
	LastName             string
	AvatarURL            string
	Activated            bool
	ActivationKey        string
	LastIP               string
}

type slack struct {
	ID          int
	Name        string
	Description string
	//Posts       []post
	Owners []int
}

//v1Cd1%*BdW73jk5DbOj#^9&E
