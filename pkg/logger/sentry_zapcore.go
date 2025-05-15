package logger

import (
	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func modifyToSentryLogger(log *zap.Logger, level zapcore.Level, component string) (*zap.Logger, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn: viper.GetString("sentry_dsn"),
	})
	if err != nil {
		// Handle the error here
	}
	cfg := zapsentry.Configuration{
		Level:             level, //when to send message to sentry
		EnableBreadcrumbs: true,  // enable sending breadcrumbs to Sentry
		BreadcrumbLevel:   level, // at what level should we sent breadcrumbs to sentry, this level can't be higher than `Level`
		Tags: map[string]string{
			"component": component,
		},
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))

	// don't use value if error was returned. Noop core will be replaced to nil soon.
	if err != nil {
		return nil, err
	}

	log = zapsentry.AttachCoreToLogger(core, log)

	// if you have web service, create a new scope somewhere in middleware to have valid breadcrumbs.
	return log.With(zapsentry.NewScope()), nil
}
