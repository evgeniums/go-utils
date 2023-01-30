package sms_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/sms"
	"github.com/evgeniums/go-backend-helpers/pkg/sms/sms_provider_factory"
	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
	"github.com/evgeniums/go-backend-helpers/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _, testBasePath, _, _ = runtime.Caller(0)
var testDir = filepath.Dir(testBasePath)

func createDb(t *testing.T, app app_context.Context) {
	test_utils.CreateDb(t, app, &sms.SmsMessage{})
}

func initSmsManager(t *testing.T) (app_context.Context, sms.SmsManager) {
	app := test_utils.InitAppContext(t, testDir, "sms_test.json")

	createDb(t, app)

	manager := sms.NewSmsManager()
	require.NoErrorf(t, manager.Init(app.Cfg(), app.Logger(), app.Validator(), &sms_provider_factory.DefaultFactory{}, "sms"), "failed to init SMS manager")

	return app, manager
}

func TestInitSmsManager(t *testing.T) {
	app, _ := initSmsManager(t)
	defer app.Close()
}

func TestSendSms(t *testing.T) {
	app, manager := initSmsManager(t)
	defer app.Close()

	user1 := user.NewUser()
	user1.InitObject()
	user1.LOGIN = "test_login1"
	user1.PHONE = "555000111"
	ctx1 := test_utils.UserOpContext(app, "TestSendSms", user1)
	message1 := "Hello world, user1"
	_, err := manager.Send(ctx1, message1, user1.PHONE)
	ctx1.Close()
	assert.NoError(t, err, "failed to send SMS")

	user2 := user.NewUser()
	user2.InitObject()
	user2.LOGIN = "test_login2"
	user2.PHONE = "999000111"
	ctx2 := test_utils.UserOpContext(app, "TestSendSms", user2)
	message2 := "Hello world, user2"
	_, err = manager.Send(ctx2, message2, user2.PHONE)
	ctx2.Close()
	assert.Error(t, err, "msut fail sending SMS")

	user3 := user.NewUser()
	user3.InitObject()
	user3.LOGIN = "test_login3"
	user3.PHONE = "990000111"
	ctx3 := test_utils.UserOpContext(app, "TestSendSms", user3)
	message3 := "Hello world, user3"
	_, err = manager.Send(ctx3, message3, user3.PHONE)
	ctx3.Close()
	assert.NoError(t, err, "failed to send SMS")
}
