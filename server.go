package main
import (
   "fmt"
   "net/http"
   "encoding/json"
   "io/ioutil"
   "errors"
   "strings"

  "github.com/auth0/go-jwt-middleware"
  "github.com/dgrijalva/jwt-go"
  "github.com/gorilla/mux"
  "github.com/rs/cors"

   "login"
   "callback"
   "api"
)

type Response struct {
   Message string
}

type Jwks struct {
   Keys []JSONWebKeys
}

type JSONWebKeys struct {
   Kty string
   Kid string
   Use string
   N   string
   E   string
   X5c []string
}


func homePage(w http.ResponseWriter, r *http.Request) {
   url := "https://dev-56plghcv.us.auth0.com/oauth/token"

   payload := strings.NewReader("{\"client_id\":\"lSwt9K86dGvblSAhp2hXbQa2TOSPt5Oa\",\"client_secret\":\"sJn6M1AUle-v3mzup1NbDVcMgbvdcOMGMGlMHukmnEl_YWQgTFPzCmV9iZn8VWgK\",\"audience\":\"http://localhost:3000/api\",\"grant_type\":\"client_credentials\"}")
 
   req, _ := http.NewRequest("POST", url, payload)
 
   req.Header.Add("content-type", "application/json")
 
   res, _ := http.DefaultClient.Do(req)
 
   defer res.Body.Close()
   body, _ := ioutil.ReadAll(res.Body)
 
   fmt.Println(res)
   fmt.Println(string(body))
   fmt.Fprint(w, string(body))
 }

 func getPemCert(token *jwt.Token) (string, error) {
   cert := ""
   resp, err := http.Get("https://cmkl-omega.auth0.com/.well-known/jwks.json")

   if err != nil {
       return cert, err
   }
   defer resp.Body.Close()

   var jwks = Jwks{}
   err = json.NewDecoder(resp.Body).Decode(&jwks)

   if err != nil {
       return cert, err
   }

   for k, _ := range jwks.Keys {
       if token.Header["kid"] == jwks.Keys[k].Kid {
           cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
       }
   }

   if cert == "" {
       err := errors.New("Unable to find appropriate key.")
       return cert, err
   }

   return cert, nil
 }

func StartServer() {
   jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
      ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
         // Verify 'aud' claim
         aud := "https://omega-next.cmkl.ac.th/"
         checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
         if !checkAud {
             return token, errors.New("Invalid audience.")
         }
         // Verify 'iss' claim
         iss := "https://cmkl-omega.auth0.com/"
         checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
         if !checkIss {
             return token, errors.New("Invalid issuer.")
         }
   
         cert, err := getPemCert(token)
         if err != nil {
             panic(err.Error())
         }
   
         result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
         return result, nil
       },
       SigningMethod: jwt.SigningMethodRS256,
})
   r := mux.NewRouter()

   r.HandleFunc("/home", homePage).Methods("GET")
   r.HandleFunc("/callback", callback.CallbackHandler).Methods("GET")
   r.HandleFunc("/login", login.LoginHandler).Methods("GET")
   r.Handle("/api/profile", jwtMiddleware.Handler(api.ProfileApiHandler)).Methods("GET")
   r.Handle("/api/enroll", jwtMiddleware.Handler(api.EnrollmentApiHandler)).Methods("GET")
   r.Handle("/api/profile", jwtMiddleware.Handler(api.UpdateProfileHandler)).Methods("POST")
   r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

   corsWrapper := cors.New(cors.Options{
      AllowedMethods: []string{"GET", "POST"},
      AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
   })

   http.ListenAndServe("0.0.0.0:3000", corsWrapper.Handler(r))
}