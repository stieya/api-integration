package activity

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var layoutTime = "2006-01-02 15:04:05"

const (
	saveQuery                    = "INSERT INTO at_api_activity (ac_user_id, ac_token, ac_api_date, ac_api_name, ac_request, ac_error_request, ac_response, ac_error_response, ac_created_by, ac_created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())"
	findAPIActivityByUserIDQuery = "SELECT id, ac_user_id, ac_token, ac_api_date, ac_api_name, ac_request, ac_error_request, ac_response, ac_error_response, ac_created_by, ac_created_at FROM at_api_activity WHERE ac_user_id = $1 ORDER BY ac_created_at DESC"
)

func NewAPIActivityRequest(userID int64, token string, date string, apiName string, request *http.Request) APIActivityRequest {
	return APIActivityRequest{
		Token:   token,
		UserID:  userID,
		Date:    date,
		APIName: apiName,
		Request: request,
	}
}

func (a *APIActivityRequest) toModel() apiActivity {
	date, _ := time.Parse(layoutTime, a.Date)
	req, errReq := formatRequest(a.Request)
	resp, errResp := formatResponse(a.response)

	return apiActivity{
		token:         a.Token,
		userID:        a.UserID,
		date:          date,
		apiName:       a.APIName,
		request:       string(req),
		errorRequest:  errReq.Error(),
		response:      string(resp),
		errorResponse: errResp.Error(),
		createdBy:     a.UserID,
		createdAt:     time.Now(),
	}
}

func (a *apiActivity) toDTO() APIActivityDTO {
	return APIActivityDTO{
		Token:         a.token,
		UserID:        a.userID,
		Date:          a.date.Format(layoutTime),
		APIName:       a.apiName,
		Request:       a.request,
		RequestError:  a.errorRequest,
		Response:      a.response,
		ResponseError: a.errorResponse,
	}
}

func (a *APIActivityRequest) Send(db interface{}) error {
	if err := a.validate(db); err != nil {
		log.Println("Failed Send APIActivity: ", err.Error())
		return err
	}

	apiActivity := a.toModel()

	go writeAPIActivityLog(a.aLogger, "[APIActivity] - Send - Request: "+fmt.Sprintf("%+v", a), a.UserID)

	if err := apiActivity.save(db); err != nil {
		log.Println("Failed Send APIActivity: ", err.Error())
		return err
	}

	go writeAPIActivityLog(a.aLogger, "[APIActivity] - Send - Response: "+fmt.Sprintf("%+v", a), a.UserID)

	return nil
}

func (a *APIActivityRequest) FindAPIActivityByUserID(db interface{}) ([]APIActivityDTO, error) {
	activities := []apiActivity{}
	activityDTOs := []APIActivityDTO{}

	if err := a.validateDBUserID(db); err != nil {
		log.Println("Failed FindAPIActivityByUserID APIActivity: ", err.Error())
		return activityDTOs, err
	}

	go writeAPIActivityLog(a.aLogger, "[APIActivity] - FindAPIActivityByUserID - UserID: "+ToString(a.UserID), a.UserID)

	activities, err := findAPIActivityByUserID(db, a.UserID)
	if err != nil {
		log.Println("Failed FindAPIActivityByUserID APIActivity: ", err.Error())
		return activityDTOs, err
	}

	for _, val := range activities {
		activityDTOs = append(activityDTOs, val.toDTO())
	}

	go writeLog(a.aLogger, "[APIActivity] - FindAPIActivityByUserID - UserID: "+ToString(a.UserID)+" - Activities: "+fmt.Sprintf("%+v", activities), a.UserID)

	return activityDTOs, nil
}

func (a *APIActivityRequest) validate(db interface{}) error {
	if db == nil {
		return fmt.Errorf("db config is null")
	}
	err := checkRequired("user_id", ToString(a.UserID))
	if err != nil {
		return err
	}
	err = checkRequired("date", ToString(a.Date))
	if err != nil {
		return err
	}
	err = checkRequired("apiName", ToString(a.APIName))
	if err != nil {
		return err
	}
	err = checkRequired("request", ToString(a.Request))
	if err != nil {
		return err
	}
	_, err = time.Parse(layoutTime, a.Date)
	if err != nil {
		return err
	}

	return nil
}

func (a *APIActivityRequest) validateDBUserID(db interface{}) error {
	if db == nil {
		return fmt.Errorf("db config is null")
	}
	err := checkRequired("user_id", ToString(a.UserID))
	if err != nil {
		return err
	}

	return nil
}

func (a *apiActivity) save(db interface{}) error {
	result, err := db.(*sqlx.DB).Exec(saveQuery, a.userID, a.token, a.date, a.apiName, a.request, a.errorRequest, a.response, a.errorResponse, a.createdBy)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("failed send apiactivity 0 rows affected")
	}

	return nil
}

func findAPIActivityByUserID(db interface{}, userID int64) ([]apiActivity, error) {
	actvities := []apiActivity{}
	query := fmt.Sprintf("%s", findAPIActivityByUserIDQuery)

	err := db.(*sqlx.DB).Select(&actvities, query, userID)
	if err != nil {
		return actvities, err
	}
	if len(actvities) == 0 {
		return actvities, fmt.Errorf("data not found")
	}

	return actvities, nil
}

func (a *APIActivityRequest) SetLogger(filepath, filename string) {
	filelog, err := NewFileLog(filepath, filename)
	if err != nil {
		log.Println("cannot load filelog user activity")
	}

	a.aLogger = filelog
}

func (a *APIActivityRequest) ClientDo(db interface{}, client http.Client, req *http.Request) (*http.Response, error) {
	// client do
	response, err := client.Do(req)

	// send apiactivity
	a.SetResponseAPI(response, err)
	go a.Send(db)

	return response, err
}

func (a *APIActivityRequest) SetResponseAPI(response *http.Response, err error) {
	if response != nil {
		a.response = response
	}

	if err != nil {
		a.errorResponse = err.Error()
	}
}

func writeAPIActivityLog(logger *FileLog, s string, userID int64) {
	go writeLog(logger, s, userID)
}
