package drive

import (
	"os"
	"log"
	"errors"
	"encoding/json"
	// "fmt"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

type JsonCreds struct {
	Type         				string `json:"type"`
	Project_id        			string `json:"project_id"`
	Private_key_id    			string `json:"private_key_id"`
	Private_key      			string `json:"private_key"`
	Client_email      			string `json:"client_email"`
	Client_id      				string `json:"client_id"`
	Auth_uri      				string `json:"auth_uri"`
	Token_uri      				string `json:"token_uri"`
	Auth_provider_x509_cert_url string `json:"auth_provider_x509_cert_url"`
	Client_x509_cert_url        string `json:"client_x509_cert_url"`
	Universe_domain         	string `json:"universe_domain"`
}

func googleAuthError(info string) error {
	return errors.New(info)
}

func GoogleAuth() *jwt.Config {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	// Navigate to your project, then see the "Credentials" page
	// under "APIs & Auth".
	// To create a service account client, click "Create new Client ID",
	// select "Service Account", and click "Create Client ID". A JSON
	// key file will then be downloaded to your computer.
	var jsonCreds JsonCreds
	jsonCreds.Type = os.Getenv("type")
	jsonCreds.Project_id = os.Getenv("project_id")
	jsonCreds.Private_key_id = os.Getenv("private_key_id")
	jsonCreds.Private_key = os.Getenv("private_key")
	jsonCreds.Client_email = os.Getenv("client_email")
	jsonCreds.Client_id = os.Getenv("client_id")
	jsonCreds.Auth_uri = os.Getenv("auth_uri")
	jsonCreds.Token_uri = os.Getenv("token_uri")
	jsonCreds.Auth_provider_x509_cert_url = os.Getenv("auth_provider_x509_cert_url")
	jsonCreds.Client_x509_cert_url = os.Getenv("client_x509_cert_url")
	jsonCreds.Universe_domain = os.Getenv("universe_domain")
    
	jsonEnv, err := json.Marshal(jsonCreds)
	jsonStr := string(jsonEnv)
	jsonStr = strings.Replace(jsonStr, "\\\\","\\",-1)

	byt := []byte(jsonStr)

	// scope := "https://www.googleapis.com/auth/drive"
	conf, err := google.JWTConfigFromJSON(byt, "https://www.googleapis.com/auth/drive")
	if err != nil {
		log.Fatal(err)
		googleAuthError("Google Authent error")
	}

	return conf
}
