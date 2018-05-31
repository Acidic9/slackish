package main

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/Acidic9/sessions"
	"github.com/asaskevich/govalidator"
	"github.com/dustinkirkland/golang-petname"
	"github.com/fatih/structs"
	"github.com/hjmodha/goDevice"
	"gopkg.in/gomail.v2"
)

func isSignedIn(session sessions.Session) bool {
	userID, valid := session.Get("id").Int64()
	if valid && userID != 0 {
		_, ok := users[int(userID)]
		if ok {
			return true
		}
	}
	return false
}

func finalizeParse(s interface{}, session sessions.Session, r *http.Request) map[string]interface{} {
	kind := reflect.TypeOf(s).Kind()
	if kind != reflect.Struct {
		panic("expected struct but got " + kind.String())
	}

	parse := structs.Map(s)
	parse["Domain"] = config.HTTP.Domain
	parse["IsSignedIn"] = isSignedIn(session)
	parse["Path"] = r.URL.Path
	parse["IsMobile"] = goDevice.GetType(r) == goDevice.MOBILE
	if _, ok := parse["Title"]; !ok {
		parse["Title"] = ""
	}

	userID, valid := session.Get("id").Int64()
	if valid {
		go updateIP(int(userID), r.RemoteAddr)
	}

	u, ok := users[int(userID)]
	if ok {
		parse["User"] = *u
	} else {
		parse["User"] = user{}
	}

	errorMessageFlashes := session.Flashes("errorMessage")
	errorMessages := make([]string, 0, len(errorMessageFlashes))
	for _, msg := range errorMessageFlashes {
		s, valid := msg.String()
		if valid && s != "" {
			errorMessages = append(errorMessages, s)
		}
	}

	successMessageFlashes := session.Flashes("successMessage")
	successMessages := make([]string, 0, len(successMessageFlashes))
	for _, msg := range successMessageFlashes {
		s, valid := msg.String()
		if valid && s != "" {
			successMessages = append(successMessages, s)
		}
	}

	parse["ErrorMessages"] = errorMessages
	parse["SuccessMessages"] = successMessages

	css := []string{
		"/css/lib/bulma.css",
		"/css/lib/font-awesome.min.css",
		"/css/lib/alertify.css",
		"/css/global.css",
		"/css/nav.css",
	}

	js := []string{
		"/js/lib/jquery-3.2.0.min.js",
		"/js/lib/alertify.js",
		"/js/nav.js",
		"/js/scripts.js",
	}

	if _, ok := parse["CSS"]; ok {
		switch reflect.TypeOf(parse["CSS"]).String() {
		case "[]string":
			parse["CSS"] = append(css, parse["CSS"].([]string)...)
		case "string":
			parse["CSS"] = append(css, parse["CSS"].(string))
		default:
			parse["CSS"] = css
		}
	} else {
		parse["CSS"] = css
	}

	if _, ok := parse["JS"]; ok {
		switch reflect.TypeOf(parse["JS"]).String() {
		case "[]string":
			parse["JS"] = append(js, parse["JS"].([]string)...)
		case "string":
			parse["JS"] = append(js, parse["JS"].(string))
		default:
			parse["JS"] = js
		}
	} else {
		parse["JS"] = js
	}

	if !parse["IsSignedIn"].(bool) {
		parse["CSS"] = append(parse["CSS"].([]string), "/css/sign-in.css")
		parse["JS"] = append(parse["JS"].([]string), "/js/sign-in.js")
	}

	return parse
}

func hashPassword(username, password, salt string) string {
	//import "crypto/md5"
	// Assume the username abc, password 123456
	h := md5.New()
	io.WriteString(h, password)

	pwmd5 := fmt.Sprintf("%x", h.Sum(nil))

	// salt1 + username + salt2 + MD5 splicing
	io.WriteString(h, salt)
	io.WriteString(h, username)
	io.WriteString(h, pwmd5)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Checks if a display name exists
func displayNameExists(name string) bool {
	var exists bool
	db.QueryRow("SELECT Exists(SELECT 1 FROM users WHERE LCase(display_name)=LCase(?))", name).Scan(&exists)
	return exists
}

func slackNameExists(name string) bool {
	var exists bool
	db.QueryRow("SELECT Exists(SELECT 1 FROM slacks WHERE LCase(name)=LCase(?));", name).Scan(&exists)
	return exists
}

func sendVerificationEmail(id int) error {
	var (
		email         sql.NullString
		displayName   sql.NullString
		activationKey sql.NullString
	)
	err := db.QueryRow("SELECT email, display_name, activation_key FROM users WHERE id=?", id).Scan(
		&email, &displayName, &activationKey)
	if err != nil {
		return fmt.Errorf("could not select user (%v): %v", id, err)
	}

	if activationKey.String == "" {
		return fmt.Errorf("activation key is empty")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "ariseyhun9@gmail.com")
	m.SetHeader("To", email.String)
	m.SetHeader("Subject", "Slackish - Account Activation")
	body := fmt.Sprintf(`<b>Welcome to Slackish!</b><br><br>You've made a new account.<br>Here are some details:<br><br><table style='font-size:16px;font-family:"Trebuchet MS",Arial,Helvetica,sans-serif;border-collapse:collapse;border-spacing:0'><tr><th style="background-color:#4caf50;color:#fff;padding:6px 12px;border:1px solid #ddd">Display Name<td style="border:1px solid #ddd;text-align:left;padding:8px">%s<tr><th style="background-color:#4caf50;color:#fff;padding:6px 12px;border:1px solid #ddd">Email<td style="border:1px solid #ddd;text-align:left;padding:8px">%s<tr><th style="background-color:#4caf50;color:#fff;padding:6px 12px;border:1px solid #ddd">Password<td style="border:1px solid #ddd;text-align:left;padding:8px">********</table><br>Your password is not stored in plain text so we cannot show it to you.<br><br>Before you can use Slackish to it's full potential, you'll need to activate your account by clicking the button below.<br><a href="`+config.HTTP.Domain+`/do/activateAccount?email=%s&key=%s"><button style="background:#3498db;background-image:-webkit-linear-gradient(top,#3498db,#2980b9);background-image:-moz-linear-gradient(top,#3498db,#2980b9);background-image:-ms-linear-gradient(top,#3498db,#2980b9);background-image:-o-linear-gradient(top,#3498db,#2980b9);background-image:linear-gradient(to bottom,#3498db,#2980b9);-webkit-border-radius:6;-moz-border-radius:6;border-radius:6px;font-family:Arial;color:#fff;font-size:20px;padding:10px 20px 10px 20px;text-decoration:none;cursor:pointer;border:none;margin:12px 4%%">Activate Account</button></a>`,
		displayName.String, email.String, email.String, activationKey.String)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 465, "ariseyhun9@gmail.com", "sorryitsprivate")
	err = d.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("could not send email to '%v': %v", strings.ToLower(email.String), err)
	}

	return nil
}

func generateActivationKey() string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, 16)

	rand.Read(bytes)

	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
}

func trimIPPort(ip string) string {
	trimmedIP := strings.Split(ip, ":")[0]
	if govalidator.IsIP(trimmedIP) {
		return trimmedIP
	}
	return ""
}

func generateDisplayName() string {
	for i := 0; i < 1000; i++ {
		displayName := petname.Generate(2, "-")
		if !displayNameExists(displayName) {
			return displayName
		}
	}
	return ""
}

func decodeGoogleIDToken(idToken string) (gplusID string, err error) {
	// An ID token is a cryptographically-signed JSON object encoded in base 64.
	// Normally, it is critical that you validate an ID token before you use it,
	// but since you are communicating directly with Google over an
	// intermediary-free HTTPS channel and using your Client Secret to
	// authenticate yourself to Google, you can be confident that the token you
	// receive really comes from Google and is valid. If your server passes the ID
	// token to other components of your app, it is extremely important that the
	// other components validate the token before using it.
	var set struct {
		Sub string
	}
	if idToken != "" {
		// Check that the padding is correct for a base64decode
		parts := strings.Split(idToken, ".")
		if len(parts) < 2 {
			return "", fmt.Errorf("Malformed ID token")
		}
		// Decode the ID token
		s := parts[1]
		switch len(s) % 4 {
		case 2:
			s += "=="
		case 3:
			s += "="
		}

		b, err := base64.URLEncoding.DecodeString(s)
		if err != nil {
			return "", fmt.Errorf("Malformed ID token: %v", err)
		}
		err = json.Unmarshal(b, &set)
		if err != nil {
			return "", fmt.Errorf("Malformed ID token: %v", err)
		}
	}
	return set.Sub, nil
}

func updateIP(id int, ip string) {
	ip = trimIPPort(ip)
	if ip == "" {
		return
	}

	db.Exec("UPDATE users SET last_ip=? WHERE id=?", ip, id)
}

func userExists(id int) bool {
	_, exists := users[id]
	return exists
}
