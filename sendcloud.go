package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

type SendCloudSender struct{}

func init() {
	Register(&SendCloudSender{})
}

//Name the name of sendcloud sender
func (*SendCloudSender) Name() string {
	return "sendcloud"
}

//SendEmail send email by sendcloud service
func (s *SendCloudSender) SendEmail(email *Email, conf *ServiceConfig) error {
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
	params := url.Values{
		"apiUser":  {conf.ApiUser},
		"apiKey":   {conf.ApiSecret},
		"from":     {from.String()},
		"fromName": {email.FromName},
		"to":       {to.String()},
		"subject":  {email.Subject},
		"html":     {email.HtmlBody},
	}

	_, err = s.sendEmailWithAttachment(conf.ApiURL, params, nil)
	if err != nil {
		return Error(err)
	}

	return nil
}

//SendSms send sms by sendcloud service
func (*SendCloudSender) SendSms(sms *Sms, conf *ServiceConfig) error {
	vars, err := json.Marshal(sms.Values)
	if err != nil {
		return Error(err)
	}
	var (
		params = url.Values{
			"smsUser":    {conf.ApiUser},
			"templateId": {sms.TemplateID},
			"vars":       {string(vars)},
		}
		signBody string
	)
	phone, err := NewPhone(sms.PhoneNumer)
	if err != nil {
		return Error(err)
	}
	if "+86" == phone.Area {
		signBody = fmt.Sprintf("%s&phone=%s&smsUser=%s&templateId=%s&vars=%s&%s",
			conf.ApiSecret, phone.SubNumber, conf.ApiUser, sms.TemplateID, string(vars), conf.ApiSecret)
		params.Add("phone", string(phone.SubNumber))
	} else {
		signBody = fmt.Sprintf("%s&msgType=2&phone=%s&smsUser=%s&templateId=%s&vars=%s&%s",
			conf.ApiSecret, phone.String(), conf.ApiUser, sms.TemplateID, string(vars), conf.ApiSecret)
		params.Add("msgType", "2")
		params.Add("phone", string(phone.String()))
	}
	params.Add("signature", MD5(signBody))
	var (
		body   = &bytes.Buffer{}
		writer = multipart.NewWriter(body)
	)
	for key, value := range params {
		_ = writer.WriteField(key, value[0])
	}
	if err := writer.Close(); err != nil {
		return Error(err)
	}
	request, err := http.NewRequest("POST", conf.ApiURL, body)
	if err != nil {
		return Error(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
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
	if result["result"] != true {
		return Error(errors.New(string(bodyByte)))
	}

	return nil
}

func (*SendCloudSender) sendEmailWithAttachment(url string, params url.Values, files *map[string][]byte) (string, error) {
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

	var err = writer.Close()
	if err != nil {
		return "", Error(err)
	}

	request, err := http.NewRequest(http.MethodPost, url, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	responseHandler, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", Error(err)
	}
	defer responseHandler.Body.Close()

	bodyByte, err := ioutil.ReadAll(responseHandler.Body)
	if err != nil {
		return string(bodyByte), Error(err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bodyByte, &result)
	if err != nil || result["result"] != true {
		return string(bodyByte), Error(err)
	}

	return string(bodyByte), nil
}
