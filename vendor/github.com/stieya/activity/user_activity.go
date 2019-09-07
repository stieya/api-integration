package activity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func NewUserActivity(url string, userID int64, token string, activity string, date string, data string, oldData string) UserActivity {
	return UserActivity{
		URL:      url,
		UserID:   userID,
		Token:    token,
		Activity: activity,
		Date:     date,
		Data:     data,
		OldData:  oldData,
	}
}

func (u *UserActivity) Send(db interface{}) error {
	// init variable
	var apiName = "UserActivity"
	var actResp = ActivityResponse{}
	var contentType = "application/json"

	if err := u.validate(); err != nil {
		log.Println("[UserActivity] - Failed validation request - ", err.Error())
		return err
	}

	go u.writeLog("[UserActivity] - Send - Request: " + fmt.Sprintf("%+v", u))

	// convert userActivity to json
	bodyRequest, err := json.Marshal(u)
	if err != nil {
		go u.writeLog("[UserActivity] - Failed Send - convert to json - Error: " + err.Error())
		log.Println("[UserActivity] - Failed convert to json - ", err.Error())
		return err
	}

	// generate http request
	client := http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", u.URL, bytes.NewBuffer(bodyRequest))
	if err != nil {
		go u.writeLog("[UserActivity] - Failed Send - new request - Error: " + err.Error())
		log.Println("[UserActivity] - Failed new request - ", err.Error())
		return err
	}

	// set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("token = %s", u.Token))

	// set api activity
	apiActivity := NewAPIActivityRequest(u.UserID, u.Token, time.Now().Format("2006-01-02 15:04:05"), apiName, req)

	// post to idm with send apiactivity
	response, err := apiActivity.ClientDo(db, client, req)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		go u.writeLog("[UserActivity] - Failed Send - Error: " + err.Error())
		log.Println("[UserActivity] - Failed timeout - ", err.Error())
		return errors.New("timeout api")
	}
	if err != nil {
		go u.writeLog("[UserActivity] - Failed Send - post - Error: " + err.Error())
		log.Println("[UserActivity] - Failed post - ", err.Error())
		return err
	}

	// read response from idm
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		go u.writeLog("[UserActivity] - Failed Send - read response body - Error: " + err.Error())
		log.Println("[UserActivity] - Failed read response body - ", err.Error())
		return err
	}

	// close response
	response.Body.Close()

	err = json.Unmarshal(body, &actResp)
	if err != nil {
		go u.writeLog("[UserActivity] - Failed Send - unmarshal response body - Error: " + err.Error())
		log.Println("[UserActivity] - Failed unmarshal response body - Error: ", err.Error())
		return err
	}
	if !actResp.Success {
		go u.writeLog("[UserActivity] - Failed Send - response success false - " + actResp.Message)
		log.Println("[UserActivity] - Failed response success false - ", actResp.Message)
		return errors.New(actResp.Message)
	}

	go u.writeLog("[UserActivity] - Send - Response: " + fmt.Sprintf("%+v", response))

	return nil
}

func (u *UserActivity) validate() error {
	err := checkRequired("user_id", ToString(u.UserID))
	if err != nil {
		return err
	}
	err = checkRequired("token", ToString(u.Token))
	if err != nil {
		return err
	}
	err = checkRequired("url", ToString(u.URL))
	if err != nil {
		return err
	}
	err = checkRequired("activity", ToString(u.Activity))
	if err != nil {
		return err
	}
	err = checkRequired("date", ToString(u.Date))
	if err != nil {
		return err
	}
	err = checkRequired("data", ToString(u.Data))
	if err != nil {
		return err
	}

	return nil
}

func (u *UserActivity) SetLogger(filepath, filename string) {
	filelog, err := NewFileLog(filepath, filename)
	if err != nil {
		log.Println("cannot load filelog user activity")
	}

	u.aLogger = filelog
}

func (u *UserActivity) writeLog(s string) {
	go writeLog(u.aLogger, s, u.UserID)
}
