package main

import (
	"fmt"
	"os"

	"github.com/Acidic9/render"
	"github.com/Acidic9/sessions"
	"github.com/go-martini/martini"
)

func main() {
	defer db.Close()

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Use(martini.Static("static"))
	store := sessions.NewCookieStore([]byte("PR6faZ9T7tbTzS4uBqrn7MkijO0sZNx55x1mjm2BqK0tqzRqHe3l2qjYiruieVdR"))
	m.Use(sessions.Sessions("slackish", store))

	// Regular GET pages
	m.Get("/", indexHandler)
	m.Get("/create", createSlackHandler)
	m.Get("\\/(?P<slack>[^/]+)\\/?$", slackHander)
	m.Get("\\/(?P<slack>[^/]+)\\/(?P<postID>\\d+)\\/?(.+)?", postHandler)

	// JSON API responses
	m.Get("/api/slacks/exists/(?P<slack>[^/]+)\\/?$", slackExistsHandler)
	m.Get("/api/displayNames/exists/(?P<displayName>[^/]+)\\/?$", displayNameExistsHandler)
	m.Get("/api/slacks/search/(?P<searchStr>[^/]+)\\/?$", slackSearchHandler)

	// JSON POST requests
	//m.Post("/do/slacks/create/(?P<slack>[^/]+)\\/?$", slackExistsHandler)
	m.Post("/do/signIn/email", signInEmailHandler)
	m.Post("/do/signIn/google", signInGoogleHandler)
	m.Get("/do/signIn/twitter", signInTwitterHandler)
	m.Get("/do/signIn/twitter/callback", signInTwitterCallbackHandler)
	m.Get("/do/signIn/github", signInGithubHandler)
	m.Post("/do/signOut", signOutHandler)
	m.Post("/do/signUp/email", signUpHandler)
	m.Get("/do/activateAccount", activateAccountHandler)
	m.Post("/do/displayName/dontAskAgain", displayNameDontAskAgainHandler)

	port := os.Getenv("PORT")
	if port == "" {
		m.RunOnAddr(fmt.Sprintf(":%d", config.HTTP.ListenPort))
		return
	}
	m.RunOnAddr(fmt.Sprintf(":%s", port))
}
