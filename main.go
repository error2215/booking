package main

import (
	"booking/templates/config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"strconv"
)

var elasticClient *elastic.Client
var Orders []config.Order

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	query := elastic.NewMatchAllQuery()

	res, _ := elasticClient.Search().
		Index(config.MAIN_INDEX).
		Pretty(true).
		Query(query).
		Size(1000).
		Do(context.Background())
	var order config.Order
	for _, hit := range res.Hits.Hits {

		order.Id, _ = strconv.Atoi(hit.Id)

		err := json.Unmarshal(hit.Source, &order)
		if err != nil {
			fmt.Println(err)
		}
		Orders = append(Orders, order)
	}
	_ = tmpl.Execute(w, Orders)
}
func init() {
	var err error

	elasticClient, err = elastic.NewClient(
		elastic.SetURL(config.ELASTIC_SERVICE_IP),
	)
	if err != nil {
		log.Println(err)
	}

	exists, err := elasticClient.IndexExists(config.MAIN_INDEX).Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if !exists {
		_, err := elasticClient.CreateIndex(config.MAIN_INDEX).Do(context.Background())
		if err != nil {
			log.Println(err)
		}
	}
	_ = elasticClient // error while elasticClient is unused

	log.Println("Configuring ended")
}

func main() {
	http.HandleFunc("/", IndexHandler)
	_ = http.ListenAndServe(":8100", nil)
}
