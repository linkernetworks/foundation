package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/logger"

	_ "bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/mongo"

	twilio "github.com/carlosdp/twiliogo"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type MessageSender interface {
	Send() error
}

type Mailgun struct {
	client mailgun.Mailgun

	domain       string
	apiKey       string
	publicApiKey string
}

type Twilio struct {
	client *twilio.TwilioClient

	accountSid string
	authToken  string
}

func NewMailgunService(mongoService *mongo.Service) *Mailgun {
	// context := mongoService.NewSession()
	// defer context.Close()
	// TODO Read config from mongo

	domain := "sandbox86ffb85f5a8d44a6bf93f5bd29fcbb79.mailgun.org"
	apiKey := "key-5edd1caa4140a3c11ee0cfd400c7c1b7"
	publicApiKey := "pubkey-0c343ddc3036d36c8027cb56d0f9da7d"

	return &Mailgun{
		domain:       domain,
		apiKey:       apiKey,
		publicApiKey: publicApiKey,
		client:       mailgun.NewMailgun(domain, apiKey, publicApiKey),
	}
}

func NewTwilioService(mongoService *mongo.Service) *Twilio {
	// context := mongoService.NewSession()
	// defer context.Close()
	// TODO Read config from mongo

	// from +19284409015
	accountSid := "ACcefaae0bfdc9accf49a7375f80217e4a"
	authToken := "5517732c5599497bda5764880bd4a45f"

	return &Twilio{
		accountSid: accountSid,
		authToken:  authToken,
		client:     twilio.NewClient(accountSid, authToken),
	}
}

func (mg *Mailgun) Send(msg *Email) error {
	message := mg.client.NewMessage(
		msg.GetSenderAddress(),
		msg.GetTitle(),
		msg.GetContent(),
		msg.GetReceiverAddress())
	resp, id, err := mg.client.Send(message)
	if err != nil {
		return err
	}
	logger.Debugf("ID: %s Resp: %s\n", id, resp)
	return nil
}

func (twlo *Twilio) Send(msg *SMS) error {
	message, err := twilio.NewMessage(
		twlo.client,
		msg.GetSenderPhoneNumber(),
		msg.GetReceiverPhoneNumber(),
		twilio.Body(msg.GetContent()))
	if err != nil {
		return err
	}
	logger.Debugln("no err:", message.Status)
	return nil
}
