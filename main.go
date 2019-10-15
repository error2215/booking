package main

import (
	login "booking-master/authorization"
	"booking-master/config"
	el "booking-master/elastic"
	"booking-master/middleware"
	"booking-master/models/order"
	"booking-master/registration"
	"booking-master/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)


func init() {
	el.Init()
	var err error

	exists, err := el.Client.IndexExists(config.MAIN_INDEX).Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	existsUser, err := el.Client.IndexExists(config.USERS_INDEX).Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if !exists {
		_, err := el.Client.CreateIndex(config.MAIN_INDEX).BodyString(config.USERS_INDEX_MAPPING).Do(context.Background())
		if err != nil {
			log.Println(err)
		}
	}

	if !existsUser {
		_, err := el.Client.CreateIndex(config.USERS_INDEX).Do(context.Background())
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Configuring ended")
	log.Println("Server started on port :", config.SERVER_PORT)
	log.Println("Elasticsearch's port:", config.ELASTICSERVICE_IP)
}

func IndexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := service.CookieStore.Get(r, "session")
		if err != nil{
			log.Error(err)
		}
		msg := service.IndexPageDataStruct{}
		deletePastRecords()
		query := elastic.NewMatchAllQuery()
		res, err := el.Client.Search().
			Index(config.MAIN_INDEX).
			Pretty(true).
			Query(query).
			Size(500).
			Sort("time",true).
			Do(context.Background())
		if err != nil {
			log.Error(err)
		}
		num := 1
		for _, hit := range res.Hits.Hits {
			data,err := hit.Source.MarshalJSON()
			if err != nil {
				fmt.Println(err)
			}
			var order2 order.Order
			order2.Number = num
			order2.Id = hit.Id
			err = json.Unmarshal(data,&order2)
			ourTime := time.Unix(order2.Time.Local().Unix() - 60 * 60 * 6,0)
			timeStr := ourTime.Format(time.RFC850)[:len(order2.Time.Format(time.RFC850)) - 7]
			order2.TimeString = timeStr
			if err != nil{
				log.Error(err)
			}
			msg.Orders = append(msg.Orders, order2)
			num++
		}

		flashes := session.Flashes()
		messages := service.GetFlashMessagesFromSession(flashes)
		msg.SuccessMessages = messages.SuccessMessages
		msg.ErrorMessages = messages.ErrorMessages
		service.ExecuteTemplateWithHeader("index", r, w, msg)
	})
}

func AddHandler() http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			_ = r.ParseForm()
			timeValue := r.Form["time"][0]
			trueTime := timeValue[:len(timeValue)-11]
			trueDate := strings.ReplaceAll(timeValue[6:], "/", ".")
			neededTime, err := time.Parse("01.02.2006 15:04", trueDate+" "+trueTime)
			if err != nil {
				log.Error(err)
			}
			order := order.Order{
				Author:  r.Form["author"][0],
				Message: r.Form["message"][0],
				Time:    neededTime.Add(time.Hour * 3),
			}

			_, err = el.Client.Index().
				Index(config.MAIN_INDEX).
				Pretty(true).
				BodyJson(order).
				Type("_doc").
				Refresh("true").
				Do(context.Background())
			if err != nil {
				log.Error(err)
			}
			http.Redirect(w, r, "/index", http.StatusTemporaryRedirect)
			return
		}
		service.ExecuteTemplateWithHeader("add",r,w, nil)
	})
}

func DeleteHandler() http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			_ = r.ParseForm()
			id := r.Form["id"][0]

			_, err := el.Client.Delete().
				Index(config.MAIN_INDEX).
				Pretty(true).
				Id(id).
				Type("_doc").
				Refresh("true").
				Do(context.Background())
			if err != nil {
				log.Error(err)
			}
			http.Redirect(w, r, "/index", http.StatusTemporaryRedirect)
			return
		}
		http.Redirect(w, r, "/index", http.StatusTemporaryRedirect)
	})
}

func main() {
	router := http.NewServeMux()
	router.Handle("/add", middleware.Adapt(AddHandler(), middleware.CheckUserRole()))
	router.Handle("/delete", middleware.Adapt(DeleteHandler(), middleware.CheckUserRole()))
	router.Handle("/login", middleware.Adapt(login.LoginHandler(), middleware.CheckUserRole()))
	router.Handle("/registration", middleware.Adapt(registration.SignInHandler(), middleware.CheckUserRole()))
	router.Handle("/", middleware.Adapt(IndexHandler(), middleware.CheckUserRole()))
	err := http.ListenAndServe(config.SERVER_PORT, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func deletePastRecords (){
	query := elastic.NewMatchAllQuery()

	res, err := el.Client.Search().
		Index(config.MAIN_INDEX).
		Pretty(true).
		Query(query).
		Size(500).
		Do(context.Background())
	if err != nil {
		log.Error(err)
	}
	for _, hit := range res.Hits.Hits {
		var order2 order.Order
		order,err := hit.Source.MarshalJSON()
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(order,&order2)
		if err != nil {
			log.Error(err)
		}
		unixRecordTime := order2.Time.Local().Unix()
		if unixRecordTime - 60 * 60  * 6 < time.Now().Local().Unix() {
			deleteRecord(hit.Id)
		}
	}
}

func deleteRecord(id string){
	_, err := el.Client.Delete().
		Index(config.MAIN_INDEX).
		Id(id).
		Type("_doc").
		Refresh("true").
		Do(context.Background())
	if err != nil {
		log.Error(err)
	}
}