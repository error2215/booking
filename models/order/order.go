package order

import "time"

type Order struct {
	Id     string           `json:"-"`
	Number int 			 `json:"-"`
	Author string        `json:"author"`
	Time   time.Time 	 `json:"time"`
	TimeString string `json:"-"`
	Message string 		 `json:"message"`
}