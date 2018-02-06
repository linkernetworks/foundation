package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"

	_ "log"

	twilio "github.com/carlosdp/twiliogo"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type MessageSender interface {
	Send(entity.Notification) error
}

type MailgunClient struct {
	client mailgun.Mailgun
}

type TwilioClient struct {
	client *twilio.TwilioClient
}

func NewMailgunService(ms entity.MailSettings) *MailgunClient {
	return &MailgunClient{
		client: mailgun.NewMailgun(ms.Mailgun.Domain, ms.Mailgun.ApiKey, ms.Mailgun.PublicApiKey),
	}
}

func NewTwilioService(tws entity.SMSSettings) *TwilioClient {
	// from +15005550006
	return &TwilioClient{
		client: twilio.NewClient(tws.Twilio.AccountSid, tws.Twilio.AuthToken),
	}
}

func (mg *MailgunClient) Send(msg entity.Notification) error {
	message := mg.client.NewMessage(msg.GetFrom(), msg.GetTitle(), msg.GetContent(), msg.GetTo()...)
	resp, id, err := mg.client.Send(message)
	if err != nil {
		return err
	}
	logger.Debugf("ID: %s Resp: %s\n", id, resp)
	return nil
}

func (twlo *TwilioClient) Send(msg entity.Notification) error {
	for _, r := range msg.GetTo() {
		message, err := twilio.NewMessage(twlo.client, msg.GetFrom(), r, twilio.Body(msg.GetContent()))
		if err != nil {
			return err
		}
		logger.Debugln("no err:", message.Status)
	}
	return nil
}
