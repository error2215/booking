package registration

import (
	"booking-master/config"
	el "booking-master/elastic"
	"booking-master/models/user"
	"booking-master/service"
	"context"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"time"

	"encoding/json"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)
func SignInHandler() http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := service.CookieStore.Get(r, "session")
		if err != nil{
			log.Error(err)
		}
		_ = session.Flashes()
		msg := service.SignInPageDataStruct{}
		if r.Method == "POST" {
			_ = r.ParseForm()
			data := make(map[string]string)

			parseValues(r.Form, data)

			validated := validate(data, session)

			if validated {
				if created := createUserInES(data); created {
					msg.SuccessMessages = append(msg.SuccessMessages,"You were signed in successfully!")
					http.Redirect(w,r,"index", http.StatusTemporaryRedirect)
					return
				}
			}
		}
		err = session.Save(r, w)
		if err != nil{
			log.Error(err)
		}

		flashes := session.Flashes()
		msg = service.SignInPageDataStruct(service.GetFlashMessagesFromSession(flashes))
		service.ExecuteTemplateWithHeader("registration", r, w, msg)
	})
}

func findUserInES(login string) bool{
	query := elastic.NewMatchAllQuery()
	res, err := el.Client.Search().
		Index(config.USERS_INDEX).
		Pretty(true).
		Type("_doc").
		Query(query).
		Size(500).
		Do(context.Background())
	if err != nil {
		log.Error("Error while finding user is ES:")
		log.Error(err)
	}
	for _, hit := range res.Hits.Hits {
		data, err := hit.Source.MarshalJSON()
		if err != nil {
			log.Error(err)
		}
		var User user.User
		err = json.Unmarshal(data,&User)
		if err != nil{
			log.Error(err)
		}
		if User.Login == login {
			return true
		}
	}
	return false
}

func createUserInES (data map[string]string) bool{

	User, err := createUserStruct(data)
	if err != nil {
		log.Error(err)
		return false
	}
	jsonData, err := json.Marshal(User)
	if err != nil {
		log.Error(err)
		return false
	}
	_, err = el.Client.Index().
				Index(config.USERS_INDEX).
				Refresh("true").
				Type("_doc").
				BodyString(string(jsonData)).
				Do(context.Background())
	if err != nil {
		log.Error("Error while adding user is ES:")
		log.Error(err)
		return false
	}
	return true
}

func validate(data map[string]string, session *sessions.Session) bool{
	good := true
	if exist := findUserInES(data["login"]); exist {
		session.AddFlash("Such login is used already, please try another")
		good = false
	}
	if data["password"] != data["password2"] {
		session.AddFlash("Passwords are not equal")
		good = false
	}
	if len(data["password"]) < 6 {
		session.AddFlash("Password is too short")
		good = false
	}
	if len(data["password"]) > 15 {
		session.AddFlash("Password is too long")
		good = false
	}
	if len(data["login"]) < 3 {
		session.AddFlash("Login is too short")
		good = false
	}
	if len(data["login"]) > 15 {
		session.AddFlash("Login is too long")
		good = false
	}
	if len(data["name"]) < 3 {
		session.AddFlash("Name is too short")
		good = false
	}
	if len(data["name"]) > 30 {
		session.AddFlash("Name is too long")
		good = false
	}
	return good
}

func parseValues(values url.Values, data map[string]string){
	data["login"] = values.Get("login")
	data["name"] = values.Get("name")
	data["password"] = values.Get("password")
	data["password2"] = values.Get("password2")
}

func createUserStruct(data map[string]string) (User user.User , err error){
	User = user.User{}
	User.Created = time.Now().Unix()
	User.Login = data["login"]
	User.Name = data["name"]
	hash,err := HashPassword(data["password"])
	if err != nil {
		return user.User{}, err
	}
	User.PassHash = hash
	User.LoginToken = "out"
	return User, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}