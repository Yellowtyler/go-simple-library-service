package main

func register(registerRequest ReqisterRequest) {

}

func login(login LoginRequest) string {
	return ""
}

func logout(token string) {

}

type ReqisterRequest struct {
	name     string
	mail     string
	password string
	role     int
}

type LoginRequest struct {
	name     string
	password string
}
