package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Acidic9/render"
	"github.com/Acidic9/sessions"
	"github.com/asaskevich/govalidator"
	"github.com/dghubble/go-twitter/twitter"
	twitterOauth1 "github.com/dghubble/oauth1"
	"github.com/go-martini/martini"
	"github.com/parnurzeal/gorequest"
)

// Regular GET handlers
func indexHandler(ren render.Render, session sessions.Session, r *http.Request) {
	parse := struct {
		CSS     string
		JS      string
		Popular []struct {
			SlackPath    string
			PostURL      string
			Title        string
			CommentCount int
			Stars        [5]int
		}
	}{
		CSS: "/css/index.css",
		JS:  "/js/index.js",
		Popular: make([]struct {
			SlackPath    string
			PostURL      string
			Title        string
			CommentCount int
			Stars        [5]int
		}, 1),
	}
	parse.Popular[0].SlackPath = "/helloworld"
	parse.Popular[0].PostURL = "/helloworld/0/first-ever-slackish-post"
	parse.Popular[0].Title = "First Ever Slack!"
	parse.Popular[0].CommentCount = 14
	parse.Popular[0].Stars[0] = 2
	parse.Popular[0].Stars[1] = 2
	parse.Popular[0].Stars[2] = 1
	parse.Popular[0].Stars[3] = 0
	parse.Popular[0].Stars[4] = 0

	ren.HTML(http.StatusOK, "index", finalizeParse(parse, session, r))
}

func createSlackHandler(params martini.Params, ren render.Render, r *http.Request, session sessions.Session) {
	parse := struct {
		CSS        string
		JS         string
		ReturnPath string
		Slack      string
	}{
		CSS:        "/css/create-slack.css",
		JS:         "/js/create-slack.js",
		ReturnPath: r.Referer(),
		Slack:      r.FormValue("slack"),
	}

	ren.HTML(http.StatusOK, "create-slack", finalizeParse(parse, session, r))
}

func slackHander(params martini.Params, ren render.Render, r *http.Request, session sessions.Session) {
	parse := struct {
		ReturnPath string
		Slack      string
	}{
		ReturnPath: r.Referer(),
		Slack:      params["slack"],
	}

	var count sql.NullInt64
	err := db.QueryRow("SELECT Count(name) FROM slacks WHERE name=?", params["slack"]).Scan(&count)
	if err != nil {
		log.Println(err)
		ren.HTML(http.StatusOK, "error-occured", finalizeParse(parse, session, r))
		return
	}

	if count.Int64 == 0 {
		ren.HTML(http.StatusOK, "slack-not-found", finalizeParse(parse, session, r))
		return
	}

	ren.HTML(http.StatusOK, "slack", finalizeParse(parse, session, r))
}

func postHandler(params martini.Params) string {
	// TODO
	return "Coming soon"
}

// JSON API handlers
func slackExistsHandler(params martini.Params, ren render.Render, r *http.Request) {
	var resp struct {
		Exists bool `json:"exists"`
	}

	resp.Exists = slackNameExists(params["slack"])

	ren.JSON(http.StatusOK, resp)
}

func displayNameExistsHandler(params martini.Params, ren render.Render, r *http.Request) {
	var resp struct {
		Exists bool `json:"exists"`
	}

	resp.Exists = displayNameExists(params["displayName"])

	ren.JSON(http.StatusOK, resp)
}

func slackSearchHandler(params martini.Params, ren render.Render, r *http.Request) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Slacks  []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"slacks"`
	}

	resp.Slacks = make([]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}, 0)

	searchStr := strings.Trim(params["searchStr"], " ")
	if !slackSearchRequirements.MatchString(searchStr) {
		resp.Success = false
		resp.Error = "Search must be 2 or more characters long"
		ren.JSON(http.StatusOK, resp)
		return
	}

	rows, err := db.Query("SELECT name, description FROM slacks WHERE name LIKE %", "%"+searchStr+"%")
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when searching slacks"
		ren.JSON(http.StatusOK, resp)
		log.Println("Error searching slacks:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var slack struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		err := rows.Scan(&slack.Name, &slack.Description)
		if err != nil {
			log.Printf("Could not scan db row: %v\n", err)
			continue
		}

		resp.Slacks = append(resp.Slacks, slack)
	}

	resp.Success = true
	ren.JSON(http.StatusOK, resp)
}

// JSON POST handlers
func signInEmailHandler(ren render.Render, r *http.Request, session sessions.Session) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	r.ParseForm()
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")
	password = hashPassword(email, password, salt)

	if !govalidator.IsEmail(email) {
		resp.Success = false
		resp.Error = "Email address is invalid"
		ren.JSON(http.StatusOK, resp)
		return
	}

	var (
		id                   sql.NullInt64
		accountType          sql.NullString
		dbEmail              sql.NullString
		displayName          sql.NullString
		displayNameGenerated sql.NullBool
		firstName            sql.NullString
		lastName             sql.NullString
		avatarURL            sql.NullString
		activated            sql.NullBool
	)

	var u user
	err := db.QueryRow("SELECT id, account_type, email, display_name, display_name_generated, first_name, last_name, avatar_url, activated FROM users WHERE email=? AND password=?", email, password).Scan(
		&id,
		&accountType,
		&dbEmail,
		&displayName,
		&displayNameGenerated,
		&firstName,
		&lastName,
		&avatarURL,
		&activated,
	)
	if err != nil {
		resp.Success = false
		resp.Error = "Incorrect username or password"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	u.ID = int(id.Int64)
	u.AccountType = accountType.String
	u.Email = dbEmail.String
	u.DisplayName = displayName.String
	u.DisplayNameGenerated = displayNameGenerated.Bool
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.AvatarURL = avatarURL.String
	u.Activated = activated.Bool

	session.Set("id", u.ID)
	users[u.ID] = &u

	resp.Success = true
	resp.Error = ""
	ren.JSON(http.StatusOK, resp)
}

func signInGoogleHandler(ren render.Render, r *http.Request, session sessions.Session) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	r.ParseForm()
	code := r.FormValue("code")

	var googleTokenResp struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		IDToken          string `json:"id_token"`
		TokenType        string `json:"token_type"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	gorequest.New().
		Post("https://www.googleapis.com/oauth2/v4/token").
		// Set("Accept", "application/json").
		Set("Content-Type", "application/x-www-form-urlencoded").
		Send(map[string]string{
			"client_id":     config.SigninAPI.Google.ClientID,
			"client_secret": config.SigninAPI.Google.ClientSecret,
			"code":          code,
			"redirect_uri":  config.HTTP.Domain,
			"grant_type":    "authorization_code",
		}).
		EndStruct(&googleTokenResp)

	if googleTokenResp.Error != "" || googleTokenResp.AccessToken == "" {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Printf("Google sign in error: %v: %v\n", googleTokenResp.Error, googleTokenResp.ErrorDescription)
		return
	}

	gplusID, err := decodeGoogleIDToken(googleTokenResp.IDToken)
	if err != nil {
		log.Fatal(err)
	}

	var profileInfo struct {
		CircledByCount int    `json:"circledByCount"`
		DisplayName    string `json:"displayName"`
		Emails         []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"emails"`
		Etag   string `json:"etag"`
		Gender string `json:"gender"`
		ID     string `json:"id"`
		Image  struct {
			IsDefault bool   `json:"isDefault"`
			URL       string `json:"url"`
		} `json:"image"`
		IsPlusUser bool   `json:"isPlusUser"`
		Kind       string `json:"kind"`
		Language   string `json:"language"`
		Name       struct {
			FamilyName string `json:"familyName"`
			GivenName  string `json:"givenName"`
		} `json:"name"`
		ObjectType  string `json:"objectType"`
		PlacesLived []struct {
			Primary bool   `json:"primary"`
			Value   string `json:"value"`
		} `json:"placesLived"`
		URL  string `json:"url"`
		Urls []struct {
			Label string `json:"label"`
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"urls"`
		Verified bool `json:"verified"`
	}

	gorequest.New().
		Get("https://www.googleapis.com/plus/v1/people/"+gplusID).
		Set("Content-Type", "application/x-www-form-urlencoded").
		Param("access_token", googleTokenResp.AccessToken).
		EndStruct(&profileInfo)

	if profileInfo.ID == "" {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println("Google sign in error: profile ID is empty")
		return
	}

	var email string
	for _, e := range profileInfo.Emails {
		if e.Type == "account" {
			email = e.Value
			break
		}
	}

	if email == "" && len(profileInfo.Emails) > 0 {
		email = profileInfo.Emails[0].Value
	}

	avatarURL := profileInfo.Image.URL
	if strings.HasPrefix(avatarURL, "https://") {
		avatarURL = strings.TrimPrefix(avatarURL, "https://")
		avatarURL = "http://" + avatarURL
	}

	displayName := strings.Replace(profileInfo.DisplayName, " ", "-", -1)
	if displayNameExists(displayName) {
		displayName = generateDisplayName()
	}

	stmt, err := db.Prepare(`INSERT INTO
		users (google_id, email, display_name, display_name_generated, first_name, last_name, avatar_url, account_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		email=?,
		first_name=?,
		last_name=?,
		avatar_url=?`)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	res, err := stmt.Exec(
		profileInfo.ID, email, displayName, true, profileInfo.Name.GivenName, profileInfo.Name.FamilyName, avatarURL, accountTypeGoogle,
		email, profileInfo.Name.GivenName, profileInfo.Name.FamilyName, avatarURL,
	)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	fmt.Println(profileInfo.ID)

	var u user
	id, err := res.LastInsertId()
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	u.ID = int(id)
	u.AccountID = profileInfo.ID
	u.AccountType = accountTypeGoogle
	u.Email = email
	u.DisplayName = displayName
	u.FirstName = profileInfo.Name.GivenName
	u.LastIP = profileInfo.Name.FamilyName
	u.AvatarURL = avatarURL

	session.Set("id", u.ID)
	users[u.ID] = &u

	resp.Success = true
	resp.Error = ""
	ren.JSON(http.StatusOK, resp)
}

func signInTwitterHandler(ren render.Render, r *http.Request, session sessions.Session) {
	token, requestURL, err := twitterConsumer.GetRequestTokenAndUrl(config.HTTP.Domain + "/do/signIn/twitter/callback")
	if err != nil {
		log.Fatal(err)
	}

	twitterTokens[token.Token] = token

	ren.Redirect(requestURL, http.StatusTemporaryRedirect)
}

func signInTwitterCallbackHandler(ren render.Render, r *http.Request, session sessions.Session) {
	values := r.URL.Query()
	verificationCode := values.Get("oauth_verifier")
	tokenKey := values.Get("oauth_token")

	token, ok := twitterTokens[tokenKey]
	if !ok {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		return
	}

	defer delete(twitterTokens, tokenKey)

	accessToken, err := twitterConsumer.AuthorizeToken(
		token,
		verificationCode,
	)
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		return
	}

	twitterConfig := twitterOauth1.NewConfig(config.SigninAPI.Twitter.APIKey, config.SigninAPI.Twitter.APISecret)
	twitterToken := twitterOauth1.NewToken(accessToken.Token, accessToken.Secret)
	httpClient := twitterConfig.Client(twitterOauth1.NoContext, twitterToken)

	c := twitter.NewClient(httpClient)

	twitterUser, _, err := c.Users.Show(&twitter.UserShowParams{
		ScreenName: accessToken.AdditionalData["screen_name"],
	})
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Printf("Error getting Twitter user info: %v\n", err)
		return
	}

	if twitterUser.IDStr == "" {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println("Twitter user ID empty")
		return
	}

	var (
		firstName string
		lastName  string
	)

	firstLastNames := strings.Split(twitterUser.Name, " ")

	firstName = firstLastNames[0]
	if len(firstLastNames) > 1 {
		lastName = firstLastNames[len(firstLastNames)-1]
	}

	stmt, err := db.Prepare(`INSERT INTO
		users (twitter_id, email, display_name, display_name_generated, first_name, last_name, avatar_url, account_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		email=?,
		first_name=?,
		last_name=?,
		avatar_url=?`)
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	displayName := strings.Replace(twitterUser.ScreenName, " ", "-", -1)
	if displayNameExists(displayName) {
		displayName = generateDisplayName()
	}

	res, err := stmt.Exec(
		twitterUser.IDStr, twitterUser.Email, displayName, true, firstName, lastName, twitterUser.ProfileImageURL, accountTypeTwitter,
		twitterUser.Email, firstName, lastName, twitterUser.ProfileImageURL,
	)
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	var u user
	id, err := res.LastInsertId()
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	u.ID = int(id)
	u.AccountID = twitterUser.IDStr
	u.AccountType = accountTypeTwitter
	u.Email = twitterUser.Email
	u.DisplayName = displayName
	u.FirstName = firstName
	u.LastIP = lastName
	u.AvatarURL = twitterUser.ProfileImageURL

	session.Set("id", u.ID)
	users[u.ID] = &u

	session.AddFlash("Signed into twitter successfully", "successMessage")
	ren.Redirect("/")
}

func signInGithubHandler(ren render.Render, r *http.Request, session sessions.Session) {
	code := r.URL.Query().Get("code")

	var githubAccessTokenResp struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}

	gorequest.New().
		Post("https://github.com/login/oauth/access_token").
		Set("Accept", "application/json").
		Send(map[string]string{
			"client_id":     config.SigninAPI.Github.ClientID,
			"client_secret": config.SigninAPI.Github.ClientSecret,
			"code":          code,
		}).
		EndStruct(&githubAccessTokenResp)

	if githubAccessTokenResp.AccessToken == "" {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println("Github access token is empty when signing user in")
		return
	}

	// Fetch user details
	var githubUserResp struct {
		Message           string `json:"message"`
		AvatarURL         string `json:"avatar_url"`
		Bio               string `json:"bio"`
		Blog              string `json:"blog"`
		Collaborators     int    `json:"collaborators"`
		Company           string `json:"company"`
		CreatedAt         string `json:"created_at"`
		DiskUsage         int    `json:"disk_usage"`
		Email             string `json:"email"`
		EventsURL         string `json:"events_url"`
		Followers         int    `json:"followers"`
		FollowersURL      string `json:"followers_url"`
		Following         int    `json:"following"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		Hireable          bool   `json:"hireable"`
		HTMLURL           string `json:"html_url"`
		ID                int    `json:"id"`
		Location          string `json:"location"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		OrganizationsURL  string `json:"organizations_url"`
		OwnedPrivateRepos int    `json:"owned_private_repos"`
		Plan              struct {
			Collaborators int    `json:"collaborators"`
			Name          string `json:"name"`
			PrivateRepos  int    `json:"private_repos"`
			Space         int    `json:"space"`
		} `json:"plan"`
		PrivateGists            int    `json:"private_gists"`
		PublicGists             int    `json:"public_gists"`
		PublicRepos             int    `json:"public_repos"`
		ReceivedEventsURL       string `json:"received_events_url"`
		ReposURL                string `json:"repos_url"`
		SiteAdmin               bool   `json:"site_admin"`
		StarredURL              string `json:"starred_url"`
		SubscriptionsURL        string `json:"subscriptions_url"`
		TotalPrivateRepos       int    `json:"total_private_repos"`
		TwoFactorAuthentication bool   `json:"two_factor_authentication"`
		Type                    string `json:"type"`
		UpdatedAt               string `json:"updated_at"`
		URL                     string `json:"url"`
	}

	gorequest.New().
		Get("https://api.github.com/user").
		Set("Accept", "application/json").
		Param("access_token", githubAccessTokenResp.AccessToken).
		EndStruct(&githubUserResp)

	if githubUserResp.Message != "" {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println("Error when signing into github:", githubUserResp.Message)
		return
	}

	stmt, err := db.Prepare(`INSERT INTO
		users (github_id, email, display_name, display_name_generated, first_name, last_name, avatar_url, account_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		email=?,
		first_name=?,
		last_name=?,
		avatar_url=?`)
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	var (
		firstName string
		lastName  string
	)
	firstLastNames := strings.Split(githubUserResp.Name, " ")

	firstName = firstLastNames[0]
	if len(firstLastNames) > 1 {
		lastName = firstLastNames[len(firstLastNames)-1]
	}

	displayName := strings.Replace(githubUserResp.Login, " ", "-", -1)
	if displayNameExists(displayName) {
		displayName = generateDisplayName()
	}

	res, err := stmt.Exec(
		githubUserResp.ID, githubUserResp.Email, displayName, true, firstName, lastName, githubUserResp.AvatarURL, accountTypeGithub,
		githubUserResp.Email, firstName, lastName, githubUserResp.AvatarURL,
	)
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	var u user
	id, err := res.LastInsertId()
	if err != nil {
		session.AddFlash("Something went wrong when siging in", "errorMessage")
		ren.Redirect("/")
		log.Println(err)
		return
	}

	u.ID = int(id)
	u.AccountID = strconv.Itoa(githubUserResp.ID)
	u.AccountType = accountTypeGithub
	u.Email = githubUserResp.Email
	u.DisplayName = displayName
	u.FirstName = firstName
	u.LastIP = lastName
	u.AvatarURL = githubUserResp.AvatarURL

	session.Set("id", u.ID)
	users[u.ID] = &u

	session.AddFlash("Signed into github successfully", "successMessage")
	ren.Redirect("/")
}

func signOutHandler(ren render.Render, session sessions.Session) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	resp.Success = true
	resp.Error = ""

	userID, valid := session.Get("id").Int64()
	if valid {
		delete(users, int(userID))
	}

	session.Clear()
	ren.JSON(http.StatusOK, resp)
}

func signUpHandler(ren render.Render, r *http.Request, session sessions.Session) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	r.ParseForm()
	displayName := r.FormValue("displayName")
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")

	if displayNameExists(displayName) {
		resp.Success = false
		resp.Error = "Display name already exists"
		ren.JSON(http.StatusOK, resp)
		return
	}

	if len(displayName) <= 4 {
		resp.Success = false
		resp.Error = "Display name is too short"
		ren.JSON(http.StatusOK, resp)
		return
	}

	if len(displayName) > 32 {
		resp.Success = false
		resp.Error = "Display name is too long"
		ren.JSON(http.StatusOK, resp)
		return
	}

	if !govalidator.IsEmail(email) {
		resp.Success = false
		resp.Error = "Email address is invalid"
		ren.JSON(http.StatusOK, resp)
		return
	}

	if len(password) < 8 {
		resp.Success = false
		resp.Error = "Password is too short"
		ren.JSON(http.StatusOK, resp)
		return
	}

	if len(password) > 64 {
		resp.Success = false
		resp.Error = "Password is too long"
		ren.JSON(http.StatusOK, resp)
		return
	}

	var exists bool
	err := db.QueryRow(
		"SELECT Exists(SELECT 1 FROM users WHERE account_type='email' AND Lower(email)=Lower(?))",
		email,
	).Scan(&exists)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	if exists {
		resp.Success = false
		resp.Error = "Email address already exists"
		ren.JSON(http.StatusOK, resp)
		return
	}

	stmt, err := db.Prepare(`INSERT INTO
		users (email, password, display_name, display_name_generated, activated, activation_key, account_type)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	u := user{
		Email:         strings.ToLower(email),
		Password:      hashPassword(email, password, salt),
		DisplayName:   displayName,
		ActivationKey: generateActivationKey(),
		AccountType:   accountTypeEmail,
	}

	res, err := stmt.Exec(u.Email, u.Password, u.DisplayName, false, false, u.ActivationKey, u.AccountType)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong when signing you in"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	u.ID = int(lastInsertID)
	go func(id int) {
		err := sendVerificationEmail(id)
		if err != nil {
			log.Println(err)
		}
	}(u.ID)

	users[u.ID] = &u

	resp.Success = true
	resp.Error = ""

	ren.JSON(http.StatusOK, resp)
}

func activateAccountHandler(r *http.Request, session sessions.Session, ren render.Render) {
	email := r.URL.Query().Get("email")
	key := r.URL.Query().Get("key")

	var activated bool
	err := db.QueryRow("SELECT activated FROM users WHERE email=Lower(?) AND activation_key=?", email, key).Scan(&activated)
	if err != nil {
		session.Set("activated", false)
		session.AddFlash("Failed to activate your account: Activation key and email don't match", "errorMessage")
		ren.Redirect("/")
		return
	}

	if activated {
		session.Set("activated", false)
		session.AddFlash("Account already activated", "errorMessage")
		ren.Redirect("/")
		return
	}

	_, err = db.Exec("UPDATE users SET activated=true WHERE email=Lower(?) AND activation_key=?", email, key)
	if err != nil {
		session.Set("activated", false)
		session.AddFlash("An error occured while activating your account", "errorMessage")
		ren.Redirect("/")
		return
	}

	userID, valid := session.Get("id").Int64()
	if valid {
		users[int(userID)].Activated = true
	}

	session.AddFlash("Account successfully activated", "successMessage")
	ren.Redirect("/")
}

func displayNameDontAskAgainHandler(r *http.Request, session sessions.Session, ren render.Render) {
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	if !isSignedIn(session) {
		resp.Success = false
		resp.Error = "You must be logged in to perform this action"
		ren.JSON(http.StatusOK, resp)
		return
	}

	userID, valid := session.Get("id").Int64()
	if !valid {
		resp.Success = false
		resp.Error = "You must be logged in to perform this action"
		ren.JSON(http.StatusOK, resp)
		return
	}

	_, err := db.Exec("UPDATE users SET display_name_generated=false WHERE id=?", userID)
	if err != nil {
		resp.Success = false
		resp.Error = "Something went wrong"
		ren.JSON(http.StatusOK, resp)
		log.Println(err)
		return
	}

	users[int(userID)].DisplayNameGenerated = false

	resp.Success = true
	resp.Error = ""
	ren.JSON(http.StatusOK, resp)
}
