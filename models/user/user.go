package user

type User struct {
	Id string `json:"-"`
	Name string `json:"name"`
	Login string `json:"login"`
	PassHash string `json:"pass_hash"`
	LoginToken string `json:"login_token"`
	Created int64 `json:"created"`
	Role int `json:"role"`
}