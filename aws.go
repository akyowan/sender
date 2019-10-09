package sender

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

type AWSSender struct{}

func init() {
	Register(&AWSSender{})
}

//Name the name of AWSSender
func (*AWSSender) Name() string {
	return "aws"
}

//SendEmail send email by aws
func (*AWSSender) SendEmail(email *Email, conf *ServiceConfig) error {
	return Error(errors.New("not support"))
}

//SendSms send sms by aws
func (*AWSSender) SendSms(sms *Sms, conf *ServiceConfig) error {
	phone, err := NewPhone(sms.PhoneNumer)
	if err != nil {
		return Error(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: credentials.NewStaticCredentials(conf.ApiUser, conf.ApiSecret, ""),
	})
	if err != nil {
		return Error(err)
	}
	var (
		snsClient = sns.New(sess)
		phoneNum  = phone.String()
		msg       = sns.PublishInput{
			Message:     &sms.Body,
			PhoneNumber: &phoneNum,
		}
	)
	_, err = snsClient.Publish(&msg)
	if err != nil {
		return Error(err)
	}

	return nil
}
