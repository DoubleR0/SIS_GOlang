package callback

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"database/sql"
	// "reflect"
	// "encoding/json"

	"github.com/coreos/go-oidc"
	_ "github.com/lib/pq"

	"app"
	"auth"
	"fmt"
)

const (
	host     = "omega-postgresql-sgp1-08776-do-user-4090996-0.db.ondigitalocean.com"
	port     = 25061
	user     = "omega_rew"
	password = "c6eqgnwwv09cxlzo"
	dbname   = "TestPool"
	sslmode = "require"
  )

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := app.Store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("state") != session.Values["state"] {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	authenticator, err := auth.NewAuthenticator()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := authenticator.Config.Exchange(context.TODO(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("no token found: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("AUTH0_CLIENT_ID"),
	}

	idToken, err := authenticator.Provider.Verifier(oidcConfig).Verify(context.TODO(), rawIDToken)

	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Getting now the userInfo
	var profile map[string]interface{}
	if err := idToken.Claims(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id_token"] = rawIDToken
	session.Values["access_token"] = token.AccessToken
	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// enc := json.NewEncoder(os.Stdout)
	// enc.Encode(profile)

	mail := strings.Split(profile["sub"].(string), "|")[1]
	firstname := profile["given_name"].(string)
	lastname := profile["family_name"].(string)
	photo := profile["picture"].(string)
	var cmkl_email string

	fmt.Println(firstname)
	fmt.Println(lastname)
	fmt.Println(mail)
	fmt.Println(profile)
	fmt.Println("")
	fmt.Println(rawIDToken)
	// fmt.Println(token.AccessToken)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", psqlInfo)
   if err != nil {
   panic(err)
   }
   defer db.Close()

   result, err := db.Query(`SELECT cmkl_email FROM student WHERE cmkl_email = $1;`, mail)
	if err != nil {
	   panic(err)
	   log.Fatal(err)
	   }
 
	   for result.Next() {
		  if err := result.Scan(&cmkl_email); err != nil {
			 log.Fatal(err)
		  }
	   }

	   if cmkl_email == "" {
		   
		sqlStatement := `INSERT INTO student (uuid, first_name, last_name, cmkl_email, photo) values($1, $2, $3, $4, $5);`

		_, err = db.Exec(sqlStatement, 0000, firstname, lastname, mail, photo)
			if err != nil {
				panic(err)
				}
		fmt.Println("passed")
	   } else {
		// sqlStatement := `UPDATE student SET tokenj = $1 WHERE cmkl_email = $2;`

		// _, err = db.Exec(sqlStatement, rawIDToken, mail)
		// if err != nil {
		// 	panic(err)
		// 	}
		fmt.Println("passed")
	   }


	http.Redirect(w, r, "/account", http.StatusSeeOther)
}
