package middleware

import (
	"booking-master/config"
	el "booking-master/elastic"
	"booking-master/models/user"
	"booking-master/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Handler func(http.Handler) http.Handler

func Adapt(h http.Handler, handlers ...Handler) http.Handler {
	for _, handler := range handlers {
		h = handler(h)
	}
	return h
}

func CheckUserRole () Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := service.CookieStore.Get(r,"auth")
			if err != nil{
				log.Error(err)
			}
			session, err := service.CookieStore.Get(r,"session")
			if err != nil{
				log.Error(err)
			}
			if session.Values["loggedIn"] == true{
				h.ServeHTTP(w, r)
				return
			}
			if cookie.Values["authorized"] != ""{
				id := cookie.Values["id"]
				userLoginToken := getLoginTokenByUserId(fmt.Sprintf("%v", id))
				loginToken := cookie.Values["loginToken"]
				if userLoginToken == loginToken{
					if err != nil{
						log.Error(err)
					}
					session.Values["loggedIn"] = true
					err = sessions.Save(r, w)
					if err != nil{
						log.Error(err)
					}
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

func getLoginTokenByUserId(id string) string{
	query := elastic.NewIdsQuery().Ids(id)
	res, err := el.Client.Search().
					Index(config.USERS_INDEX).
					Type("_doc").
					Query(query).
					Do(context.Background())
	if err != nil{
		log.Error(err)
	}
	if res.Hits.TotalHits > 0 {
		for _, hit := range res.Hits.Hits {
			data,err :=hit.Source.MarshalJSON()
			if err != nil{
				log.Error(err)
			}
			var User user.User
			err = json.Unmarshal(data,&User)
			if err != nil {
				log.Error(err)
			}
			return User.LoginToken
		}
	}
	return "out"
}