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

func NewCustomerActivity(url string, customerID int64, userID int64, token string, activity string, date string, data string, oldData string) CustomerActivity {
	return CustomerActivity{
		URL:        url,
		CustomerID: customerID,
		UserID:     userID,
		Token:      token,
		Activity:   activity,
		Date:       date,
		Data:       data,
		OldData:    oldData,
	}
}

func (c *CustomerActivity) Send(db interface{}) error {
	// init variable
	var apiName = "CustomerActivity"
	var actResp = ActivityResponse{}
	var contentType = "application/json"

	if err := c.validate(); err != nil {
		log.Println("[CustomerActivity] - Failed validation request - ", err.Error())
		return err
	}

	go c.writeLog("[CustomerActivity] - Send - Request: " + fmt.Sprintf("%+v", c))

	// convert customerActivity to json
	bodyRequest, err := json.Marshal(c)
	if err != nil {
		go c.writeLog("[CustomerActivity] - Failed Send - convert to json - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed convert to json - ", err.Error())
		return err
	}

	// generate http request
	client := http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(bodyRequest))
	if err != nil {
		go c.writeLog("[CustomerActivity] - Failed Send - new request - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed new request - ", err.Error())
		return err
	}

	// set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("token = %s", c.Token))

	// set api activity
	apiActivity := NewAPIActivityRequest(c.UserID, c.Token, time.Now().Format("2006-01-02 15:04:05"), apiName, req)

	// post to idm with send apiactivity
	response, err := apiActivity.ClientDo(db, client, req)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		go c.writeLog("[CustomerActivity] - Failed Send - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed timeout - ", err.Error())
		return errors.New("timeout api")
	}
	if err != nil {
		go c.writeLog("[CustomerActivity] - Failed Send - post - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed post - ", err.Error())
		return err
	}

	// read response from idm
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		go c.writeLog("[CustomerActivity] - Failed Send - read response body - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed read response body - ", err.Error())
		return err
	}

	// close response
	response.Body.Close()

	err = json.Unmarshal(body, &actResp)
	if err != nil {
		go c.writeLog("[CustomerActivity] - Failed Send - unmarshal response body - Error: " + err.Error())
		log.Println("[CustomerActivity] - Failed unmarshal response body - Error: ", err.Error())
		return err
	}
	if !actResp.Success {
		go c.writeLog("[CustomerActivity] - Failed Send - response success false - " + actResp.Message)
		log.Println("[CustomerActivity] - Failed response success false - ", actResp.Message)
		return errors.New(actResp.Message)
	}

	go c.writeLog("[CustomerActivity] - Send - Response: " + fmt.Sprintf("%+v", response))

	return nil
}

func (c *CustomerActivity) validate() error {
	err := checkRequired("customer_id", ToString(c.CustomerID))
	if err != nil {
		return err
	}
	err = checkRequired("user_id", ToString(c.UserID))
	if err != nil {
		return err
	}
	err = checkRequired("token", ToString(c.Token))
	if err != nil {
		return err
	}
	err = checkRequired("url", ToString(c.URL))
	if err != nil {
		return err
	}
	err = checkRequired("activity", ToString(c.Activity))
	if err != nil {
		return err
	}
	err = checkRequired("date", ToString(c.Date))
	if err != nil {
		return err
	}
	err = checkRequired("data", ToString(c.Data))
	if err != nil {
		return err
	}

	return nil
}

func (c *CustomerActivity) SetLogger(filepath, filename string) {
	filelog, err := NewFileLog(filepath, filename)
	if err != nil {
		log.Println("cannot load filelog user activity")
	}

	c.aLogger = filelog
}

func (c *CustomerActivity) writeLog(s string) {
	go writeLog(c.aLogger, s, c.UserID)
}
