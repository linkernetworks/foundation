package messenger

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	// "gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"

	_ "log"
	"testing"
)

func TestMailgunNewService(t *testing.T) {
	toUser := "develop@linkernetworks.com"
	toUser2 := "develop@linkernetworks.com"
	fromUser := "noreply@linkernetworks.com"

	title := "Hello from Mailgun"
	content := `This is a long content. Lorem ipsum dolor sit amet, consectetuer adipiscing
elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis
dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque
eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo,  fringilla vel,
aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet  a, venenatis vitae,
justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus
elementum semper nisi. Aenean vulputate eleifend tellus. Aeneanleo ligula, porttitor eu,
consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a,
tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet.
Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam
rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet
adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id,
lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis
faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed
fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales,
augue velit cursus nunc,`

	e := NewEmail(title, content, fromUser, toUser, toUser2)
	assert.NotNil(t, e)

	mailgunSetting := entity.MailSettings{
		Mailgun: entity.Mailgun{
			Domain:       "sandbox86ffb85f5a8d44a6bf93f5bd29fcbb79.mailgun.org",
			ApiKey:       "key-5edd1caa4140a3c11ee0cfd400c7c1b7",
			PublicApiKey: "pubkey-0c343ddc3036d36c8027cb56d0f9da7d",
		},
	}

	mg := NewMailgunService(mailgunSetting)
	err := mg.Send(e)
	assert.NoError(t, err)
}

func TestTwilioNewService(t *testing.T) {
	toUser := "+886952301269"
	toUser2 := "+886952301269"
	fromUser := "+15005550006"

	content := "Hello from Twillio. This is the test case message for tesing TestTwilioNewService"

	sms := NewSMS(content, fromUser, toUser, toUser2)
	assert.NotNil(t, sms)

	twilioSetting := entity.SMSSettings{
		Twilio: entity.Twilio{
			AccountSid: "ACa840cade3f49c7fed9ee56ecea044a4b",
			AuthToken:  "d9e40c67467eafab94bd7b6603bfa7b4",
		},
	}

	twlo := NewTwilioService(twilioSetting)
	err := twlo.Send(sms)
	assert.NoError(t, err)
}
