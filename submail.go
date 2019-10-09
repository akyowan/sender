package sender

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type SubmailSender struct{}

func init() {
	Register(&SubmailSender{})
}

//Name the name of submail sender
func (*SubmailSender) Name() string {
	return "submail"
}

//SendEmail send email by submail
func (s *SubmailSender) SendEmail(email *Email, conf *ServiceConfig) error {
	if conf.ApiUser == "" {
		return Error(errors.New("config[ApiUser] is nil"))
	}
	if conf.ApiSecret == "" {
		return Error(errors.New("config[ApiSecret] is nil"))
	}
	if email.Subject == "" {
		return Error(errors.New("email subject is nil"))
	}
	if email.HtmlBody == "" {
		return Error(errors.New("email html body is nil"))
	}
	from, err := NewEmail(email.From)
	if err != nil {
		return Error(err)
	}
	to, err := NewEmail(email.To)
	if err != nil {
		return Error(err)
	}
	var (
		params = url.Values{
			"appid":     {conf.ApiUser},
			"to":        {to.String()},
			"from":      {from.String()},
			"subject":   {email.Subject},
			"html":      {email.HtmlBody},
			"signature": {conf.ApiSecret},
		}
	)

	_, err = s.sendEmailWithAttachment(conf.ApiURL, params, nil)
	if err != nil {
		return Error(err)
	}

	return nil
}

//SendSms send email by submail
func (*SubmailSender) SendSms(sms *Sms, conf *ServiceConfig) error {
	phone, err := NewPhone(sms.PhoneNumer)
	if err != nil {
		return Error(err)
	}
	var (
		params = url.Values{
			"appid":     {conf.ApiUser},
			"to":        {phone.String()},
			"signature": {conf.ApiSecret},
			"content":   {sms.Body},
		}
	)

	request, err := http.NewRequest(http.MethodPost, conf.ApiURL, bytes.NewReader([]byte(params.Encode())))
	if err != nil {
		return Error(err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	responseHandler, err := http.DefaultClient.Do(request)
	if err != nil {
		return Error(err)
	}
	defer responseHandler.Body.Close()

	bodyByte, err := ioutil.ReadAll(responseHandler.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(bodyByte, &result); err != nil {
		return Error(err)
	}
	if result["status"] != "success" {
		return Error(errors.New(string(bodyByte)))
	}

	return nil
}

func (*SubmailSender) sendEmailWithAttachment(url string, params url.Values, files *map[string][]byte) (string, error) {
	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
	)
	if files != nil {
		for k, v := range *files {
			fileWriter, err := writer.CreateFormFile("attachments", k)
			if err != nil {
				return "", Error(err)
			}
			if _, err := fileWriter.Write(v); err != nil {
				return "", Error(err)
			}
		}
	}
	for key, value := range params {
		if err := writer.WriteField(key, value[0]); err != nil {
			return "", Error(err)
		}
	}
	if err := writer.Close(); err != nil {
		return "", Error(err)
	}
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", Error(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	responseHandler, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", Error(err)
	}
	defer responseHandler.Body.Close()

	bodyByte, err := ioutil.ReadAll(responseHandler.Body)
	if err != nil {
		return "", Error(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyByte, &result); err != nil {
		return string(bodyByte), Error(err)
	}
	if result["status"] != "success" {
		return string(bodyByte), Error(errors.New(string(bodyByte)))
	}
	return string(bodyByte), nil
}
