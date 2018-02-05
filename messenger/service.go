package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	"bitbucket.org/linkernetworks/aurora/src/logger"

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
	// ms.Mailgun.ApiKey
	// ms.Mailgun.Domain
	// ms.Mailgun.PublicApiKey
	domain := "sandbox86ffb85f5a8d44a6bf93f5bd29fcbb79.mailgun.org"
	apiKey := "key-5edd1caa4140a3c11ee0cfd400c7c1b7"
	publicApiKey := "pubkey-0c343ddc3036d36c8027cb56d0f9da7d"
	return &MailgunClient{
		client: mailgun.NewMailgun(domain, apiKey, publicApiKey),
	}
}

func NewTwilioService(tws entity.SMSSettings) *TwilioClient {
	// from +19284409015
	// tws.Twilio.AccountSid
	// tws.Twilio.AuthToken
	accountSid := "ACcefaae0bfdc9accf49a7375f80217e4a"
	authToken := "5517732c5599497bda5764880bd4a45f"
	return &TwilioClient{
		client: twilio.NewClient(accountSid, authToken),
	}
}

func (mg *MailgunClient) Send(msg entity.Notification) error {
	message := mg.client.NewMessage(msg.GetFrom(), msg.GetTitle(), msg.GetContent(), msg.GetTo())
	resp, id, err := mg.client.Send(message)
	if err != nil {
		return err
	}
	logger.Debugf("ID: %s Resp: %s\n", id, resp)
	return nil
}

func (twlo *TwilioClient) Send(msg entity.Notification) error {
	message, err := twilio.NewMessage(twlo.client, msg.GetFrom(), msg.GetTo(), twilio.Body(msg.GetContent()))
	if err != nil {
		return err
	}
	logger.Debugln("no err:", message.Status)
	return nil
}
