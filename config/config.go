package config

const (
	ELASTICSERVICE_IP = "http://localhost:9200"

	MAIN_INDEX         = "booking_index"
	USERS_INDEX = "users_index"

	SERVER_PORT 	   = ":8100"
	SERVER_DOMAIN = "localhost"

	SESSION_KEY = "Du2kaQK797dFRv7sPKCRcIefKhPdIj6l"
)

const (
	USERS_INDEX_MAPPING = `{
"mappings": {
    "_doc": { 
      "properties": { 
        "name":{ "type": "keyword"  }, 
        "login":{ "type": "keyword"  }, 
        "pass_hash":{ "type": "keyword" },  
        "created":{"type":"long"},
"login_token" :{"type": "keyword"}
      }
    }
  }
}`
)