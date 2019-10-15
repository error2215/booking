package authorization

import (
	"booking-master/config"
	el "booking-master/elastic"
	"booking-master/models/user"
	"booking-master/service"
	"context"
	"encoding/json"
	"github.com/gorilla/sessions"
	"github.com/olivere/elastic"

	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"

	log "github.com/sirupsen/logrus"
)
func LoginHandler() http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := service.CookieStore.Get(r, "session")
		if err != nil{
			log.Error(err)
		}
		_ = session.Flashes()
		msg := service.LoginPageDataStruct{}
		if r.Method == "POST" {
			_ = r.ParseForm()

			login := r.Form["login"][0]
			password := r.Form["password"][0]
			checkVal := r.Form["check"]
			check := ""
			if len(checkVal) > 0{
				check = checkVal[0]
			}

			query := elastic.NewMatchQuery("login", login)
			res, err := el.Client.Search().
				Type("_doc").
				Index(config.USERS_INDEX).
				Size(500).
				Query(query).
				Do(context.Background())
			if err != nil {
				log.Error("Error while finding user is ES")
				log.Error(err)
				msg.ErrorMessages = append(msg.ErrorMessages,"Login or password is not correct")
				service.ExecuteTemplateWithHeader("login", r, w, msg)
				return
			}
			if res.Hits.TotalHits > 0 {
				for _, hit := range res.Hits.Hits {
					data, err := hit.Source.MarshalJSON()
					if err != nil {
						log.Error(err)
					}
					var User user.User
					User.Id = hit.Id
					err = json.Unmarshal(data, &User)
					if err != nil {
						log.Error(err)
					}
					if equals := CheckPasswordHash(password, User.PassHash); equals {
						User.LoginToken = AuthUser(r ,w, check, User.Id)
						updateUserToken(User)
						msg.SuccessMessages = append(msg.SuccessMessages,"You're in!")
						http.Redirect(w,r,"index", http.StatusTemporaryRedirect)
						return
					} else {
						msg.ErrorMessages = append(msg.ErrorMessages,"Login or password is not correct")
						service.ExecuteTemplateWithHeader("login", r, w,msg )
						return
					}
				}
			} else {
				msg.ErrorMessages = append(msg.ErrorMessages,"Login or password is not correct")
				service.ExecuteTemplateWithHeader("login", r, w,msg )
				return
			}
		}
		flashes := session.Flashes()
		msg = service.LoginPageDataStruct(service.GetFlashMessagesFromSession(flashes))
		service.ExecuteTemplateWithHeader("login", r, w,msg)
	})
}
func AuthUser (r *http.Request, w http.ResponseWriter, rememberMe string, id string) string{
	maxAge := 0
	if rememberMe == "true" {
		maxAge = 60 * 60 * 24 * 30
	}
	loginToken := randStr(36)
	// login token to
	cookie, err := service.CookieStore.Get(r, "auth")
	if err != nil {
		log.Error(err)
	}
	cookie.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
	}
	cookie.Values["loginToken"] = loginToken
	cookie.Values["authorized"] = "true"
	cookie.Values["id"] = id
	err = sessions.Save(r, w)
	if err != nil{
		log.Error(err)
	}

	return loginToken
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func updateUserToken(User user.User){
	jsonBytes, err := json.Marshal(User)
	if err != nil {
		log.Error(err)
	}
	_, err = el.Client.Index().
				Index(config.USERS_INDEX).
				Id(User.Id).
				BodyJson(string(jsonBytes)).
				Type("_doc").
				Refresh("true").
				Do(context.Background())
	if err != nil{
		log.Error(err)
	}
}