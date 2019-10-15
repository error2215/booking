package elastic

import (
	"booking-master/config"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

var Client *elastic.Client


func Init()  {
	var err error
	Client, err = elastic.NewClient(
		elastic.SetURL(config.ELASTICSERVICE_IP),
	)
	if err != nil {
		log.Println(err)
	}
	_ = Client // error while ElasticClient is unused
}