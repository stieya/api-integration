package apiintegration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/spartaut/activity"
)

type APIIntegration struct {
	UserID      int64
	Token       string
	APICode     string
	Method      string
	ContentType string
	Host        string
	Timeout     int
	ObjReq      interface{}
	Headers     map[string]string
	IsLocalAPI  bool
	Logger      *log.Logger
}

func (a *APIIntegration) Send(db interface{}) (interface{}, error) {
	var resp interface{}

	// generate http request
	client := http.Client{Timeout: a.validateTimeout() * time.Second}
	req, err := http.NewRequest(a.Method, a.Host, a.generateBody())
	if err != nil {
		go a.writeLog("[" + a.APICode + "] - Failed Send - new request - Error: " + err.Error())
		log.Println("[", a.APICode, "] - Failed new request - ", err.Error())
		return nil, err
	}

	// set headers
	a.generateHeaders(req)

	// set api activity
	apiActivity := activity.NewAPIActivityRequest(a.UserID, a.Token, time.Now().Format("2006-01-02 15:04:05"), a.APICode, req)

	// post to idm with send apiactivity
	response, err := apiActivity.ClientDo(db, client, req)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		go a.writeLog("[" + a.APICode + "] - Failed Send - Error: " + err.Error())
		log.Println("[", a.APICode, "] - Failed timeout - ", err.Error())
		return nil, errors.New("timeout api")
	}
	if err != nil {
		go a.writeLog("[" + a.APICode + "] - Failed Send - post - Error: " + err.Error())
		log.Println("[", a.APICode, "] - Failed post - ", err.Error())
		return nil, err
	}

	// read response from idm
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		go a.writeLog("[" + a.APICode + "] - Failed Send - read response body - Error: " + err.Error())
		log.Println("[", a.APICode, "] - Failed read response body - ", err.Error())
		return nil, err
	}

	// close response
	response.Body.Close()

	err = json.Unmarshal(body, &resp)
	if err != nil {
		go a.writeLog("[" + a.APICode + "] - Failed Send - unmarshal response body - Error: " + err.Error())
		log.Println("[", a.APICode, "] - Failed unmarshal response body - Error: ", err.Error())
		return nil, err
	}

	return resp, nil
}

func (a *APIIntegration) validateTimeout() time.Duration {
	defaultTimeout := 5
	if a.Timeout == 0 {
		return time.Duration(defaultTimeout)
	}
	return time.Duration(a.Timeout)
}

func (a *APIIntegration) generateHeaders(req *http.Request) {
	req.Header.Set("Content-Type", a.ContentType)
	for key, value := range a.Headers {
		req.Header.Set(key, value)
	}
	if a.IsLocalAPI {
		req.Header.Set("Authorization", fmt.Sprintf("token = %s", a.Token))
	}
}

func (a *APIIntegration) generateBody() io.Reader {
	if a.ObjReq != nil {
		bodyRequest, err := json.Marshal(a.ObjReq)
		if err == nil {
			return bytes.NewBuffer(bodyRequest)
		}
	}
	return nil
}

func (a *APIIntegration) writeLog(s string) {
	if a.Logger == nil {
		return
	}

	str := fmt.Sprintf("[%d] - %s", a.UserID, s)
	if a.Logger != nil {
		go a.Logger.Printf("[%s] %q", EncodeMD5(ToString(str)), str)
	}
}
