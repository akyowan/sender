package sender

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type MXCSender struct{}

func init() {
	Register(&MXCSender{})
}

// mxc response
type mxcResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

//Name the mxc sender name
func (*MXCSender) Name() string {
	return "mxc"
}

//SendEmail send email by mxc
func (*MXCSender) SendEmail(email *Email, conf *ServiceConfig) error {
	if conf.ApiURL == "" {
		return Error(errors.New("api url is nil"))
	}
	if email.Uid == "" {
		return Error(errors.New("no uid"))
	}
	if email.Subject == "" {
		return Error(errors.New("no subject"))
	}
	if email.HtmlBody == "" {
		return Error(errors.New("no html body"))
	}

	params := url.Values{}
	params.Add("uid", email.Uid)
	params.Add("title", email.Subject)
	params.Add("content", email.HtmlBody)
	response, err := http.PostForm(conf.ApiURL, params)
	if err != nil {
		return Error(err)
	}
	defer response.Body.Close()

	var resp mxcResponse
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return Error(err)
	}

	if !resp.Success {
		return Error(errors.New(resp.Message))
	}

	return nil
}

// SendSms send sms by mxc
func (*MXCSender) SendSms(sms *Sms, conf *ServiceConfig) error {
	if conf.ApiURL == "" {
		return Error(errors.New("api url is nil"))
	}
	if sms.Uid == "" {
		return Error(errors.New("no uid"))
	}
	if sms.Body == "" {
		return Error(errors.New("no body"))
	}

	params := url.Values{}
	params.Add("uid", sms.Uid)
	params.Add("content", sms.Body)
	response, err := http.PostForm(conf.ApiURL, params)
	if err != nil {
		return Error(err)
	}
	defer response.Body.Close()

	var resp mxcResponse
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return Error(err)
	}

	if !resp.Success {
		return Error(errors.New(resp.Message))
	}

	return nil
}
