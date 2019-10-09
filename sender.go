package sender

import (
	"fmt"
	"github.com/pkg/errors"
)

//ServiceConfig sender service config
type ServiceConfig struct {
	Region    string // Only for aws
	ApiUser   string // ApiUser(sendcloud), AccessKey(aws), appID(submail)
	ApiSecret string // ApiKey(sendcloud/sendgrid/submail) AccessSecret(aws)
	ApiURL    string
}

//Email email object
type Email struct {
	Service  string `json:"service"`
	Uid      string `json:"uid"` // only for mxc
	From     string `json:"from"`
	FromName string `json:"from_name"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HtmlBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

//Sms sms object
type Sms struct {
	Service    string      `json:"service"`
	Uid        string      `json:"uid"` // only for mxc
	PhoneNumer string      `json:"phone_numer"`
	TemplateID string      `json:"template_id"`
	Values     interface{} `json:"source"`
	Body       string      `json:"body"`
}

//Sender sender interface
type Sender interface {
	Name() string
	SendEmail(email *Email, conf *ServiceConfig) error
	SendSms(sms *Sms, conf *ServiceConfig) error
}

var (
	scf     func(service, phone string) *ServiceConfig
	ecf     func(service string) *ServiceConfig
	senders = map[string]Sender{}
)

const (
	SENDCLOUD = "sendcloud" // https://sendcloud.sohu.com/ (email&sms)
	SUBMAIL   = "submail"   // https://www.mysubmail.com/ (email&sms)
	NEC       = "nec"       // http://dun.163.com (sms)
	AWS       = "aws"       // https://docs.aws.amazon.com/zh_cn/sns/latest/dg/sns-mobile-phone-number-as-subscriber.html (sms)
	SENDGRID  = "sendgrid"  // https://sendgrid.com (email)
	MIAOSAI   = "miaosai"   // https://www.shmiaosai.com (sms)
	MXC       = "mxc"       // MXC api (email/sms)
)

//SetConfigFunc Set get service config func
func SetSmsConfigFunc(f func(service, phone string) *ServiceConfig) {
	scf = f
}

//SetEmailConfigFunc Set get service config func
func SetEmailConfigFunc(f func(service string) *ServiceConfig) {
	ecf = f
}

//RegisterSender register an sender
func Register(sender Sender) {
	senders[sender.Name()] = sender
}

//SendEmail send email
func SendEmail(email *Email) error {
	if email == nil {
		panic("email pointer is nil")
	}
	if ecf == nil {
		return Error(errors.New("Not set config func"))
	}
	conf := ecf(email.Service)
	if conf == nil {
		return Error(errors.Errorf("service[%s] config is nil", email.Service))
	}
	sender, ok := senders[email.Service]
	if !ok {
		return Error(errors.Errorf("sender[%s] not register", email.Service))
	}
	return Error(sender.SendEmail(email, conf))
}

//SendSms send sms
func SendSms(sms *Sms) error {
	if sms == nil {
		panic("sms pointer is nil")
	}
	if scf == nil {
		return Error(errors.New("Not set config func"))
	}
	conf := scf(sms.Service, sms.PhoneNumer)
	if conf == nil {
		return Error(errors.New(fmt.Sprintf("service[%s] config is nil", sms.Service)))
	}

	sender, ok := senders[sms.Service]
	if !ok {
		return Error(errors.Errorf("sender[%s] not register", sms.Service))
	}
	return Error(sender.SendSms(sms, conf))
}
