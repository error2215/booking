package service

import (
	"booking-master/config"
	"booking-master/models/order"
	"fmt"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

var CookieStore = sessions.NewCookieStore([]byte(config.SESSION_KEY))

type IndexPageDataStruct struct {
	ErrorMessages []string
	SuccessMessages []string
	Orders [] order.Order
}

type LoginPageDataStruct struct {
	ErrorMessages []string
	SuccessMessages []string
}

type SignInPageDataStruct struct {
	ErrorMessages []string
	SuccessMessages []string
}

type FlashMessagesStruct struct {
	ErrorMessages []string
	SuccessMessages []string
}

func ExecuteTemplateWithHeader (path string,r *http.Request, w http.ResponseWriter, data interface{}){
	filenames := []string{"templates/" + path +".html"}
	session, err := CookieStore.Get(r,"session")
	if err != nil{
		log.Error(err)
	}
	logged := false
	value := session.Values["logged"]
	log.Info(value)
	if len(fmt.Sprintf("%v", value)) > 0 && value == true{
		logged = true
	}
	log.Info(logged)
	if logged {
		filenames = append(filenames, "templates/header_logged.html")
	}else{
		filenames = append(filenames, "templates/header.html")
	}

	var tmpl = template.Must(template.ParseFiles(filenames...))
	_ = tmpl.ExecuteTemplate(w, path, data)
}


func GetFlashMessagesFromSession(flashes[]interface{}) FlashMessagesStruct{
	result := FlashMessagesStruct{}
	if len(flashes) > 0 {
		for _, value := range flashes{
			message := fmt.Sprintf("%v", value)
			if message[:len(message)-7] == "success" {
				result.SuccessMessages = append(result.SuccessMessages,message[7:])
			}else{
				result.ErrorMessages = append(result.ErrorMessages,message)
			}
		}
	}
	return result
}
