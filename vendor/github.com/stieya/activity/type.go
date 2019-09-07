package activity

import (
	"log"
	"net/http"
	"time"
)

type apiActivity struct {
	id            int64     `db:"ac_id"`
	token         string    `db:"ac_token"`
	userID        int64     `db:"ac_user_id"`
	date          time.Time `db:"ac_api_date"`
	apiName       string    `db:"ac_api_name"`
	request       string    `db:"ac_request"`
	response      string    `db:"ac_response"`
	errorRequest  string    `db:"ac_error_request"`
	errorResponse string    `db:"ac_error_response"`
	createdAt     time.Time `db:"ac_created_at"`
	createdBy     int64     `db:"ac_created_by"`
}

type APIActivityRequest struct {
	Token         string        `json:"token"`
	UserID        int64         `json:"user_id"`
	Date          string        `json:"date"`
	APIName       string        `json:"api_name"`
	Request       *http.Request `json:"request"`
	response      *http.Response
	errorResponse string
	aLogger       *FileLog
}

type APIActivityDTO struct {
	Token         string `json:"token"`
	UserID        int64  `json:"user_id"`
	Date          string `json:"date"`
	APIName       string `json:"api_name"`
	Request       string `json:"request"`
	RequestError  string `json:"request_error"`
	Response      string `json:"response"`
	ResponseError string `json:"response_error"`
}

type UserActivity struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Activity string `json:"activity"`
	Date     string `json:"activity_date"`
	Data     string `json:"data"`
	OldData  string `json:"old_data"`
	aLogger  *FileLog
}

type CustomerActivity struct {
	URL        string `json:"url"`
	Token      string `json:"token"`
	UserID     int64  `json:"user_id"`
	CustomerID int64  `json:"customer_id"`
	Activity   string `json:"activity"`
	Date       string `json:"activity_date"`
	Data       string `json:"data"`
	OldData    string `json:"old_data"`
	aLogger    *FileLog
}

type ActivityResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type FileLog struct {
	Conn     *log.Logger
	FilePath string
	FileName string
}

type Logger interface {
	GenerateCode(string) string
	Write(string, interface{})
}
