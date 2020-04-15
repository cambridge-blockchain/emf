package logger

import (
	"fmt"
	"log/syslog"
	"os"
	"time"

	elastic_logrus "github.com/interactive-solutions/go-logrus-elasticsearch"
	"github.com/labstack/echo/v4"
	emiddleware "github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"

	"github.com/cambridge-blockchain/emf/configurer"
)

// DefaultLoggingMiddleware exposes the default logging middleware for the platform
var DefaultLoggingMiddleware = emiddleware.LoggerWithConfig(
	emiddleware.LoggerConfig{
		Format: `[${time_rfc3339}] ${status} ${method} ${uri} | request_id=${header:` +
			echo.HeaderXRequestID + `} user_agent="${user_agent}" remote_ip=${remote_ip} ` +
			`bytes_in=${bytes_in} bytes_out=${bytes_out} latency=${latency_human}` + "\n",
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/info" || c.Path() == "/metrics"
		},
	},
)

func indexNameFunc(bc configurer.BuildConfig) elastic_logrus.IndexNameFunc {
	return func() string {
		return fmt.Sprintf("%s-%s-%s",
			bc.Component,
			bc.Version,
			time.Now().Format("2006-01-02"),
		)
	}
}

// New : Creates a default echo logger object, properly configured
func New(conf configurer.Config, bc configurer.BuildConfig) echo.Logger {
	var err error

	// ***********************************************
	// * Set up Logger
	// ***********************************************

	var isElasticSearch = conf.GetBool("logging.elasticsearch")
	var isSyslog = conf.GetBool("logging.syslog")

	if !isElasticSearch && !isSyslog {
		var l = elog.New(bc.Component)
		l.EnableColor()
		l.SetHeader(`[${time_rfc3339}] ${level} ${prefix} @ ${short_file}:${line} |`)
		l.SetLevel(elog.DEBUG)
		return l
	}

	var url = conf.GetString("logging.endpoint")
	var syslogURL = conf.GetString("logging.syslog_endpoint")

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	if isElasticSearch {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})

		var client *elastic.Client
		if client, err = elastic.NewClient(elastic.SetURL(url)); err != nil {
			panic(err)
		}

		var elkHook *elastic_logrus.ElasticSearchHook
		if elkHook, err = elastic_logrus.NewElasticHook(
			client,
			fmt.Sprintf("%s-%s",
				bc.Component,
				bc.Version,
			),
			logrus.DebugLevel,
			indexNameFunc(bc),
			time.Second*base10,
		); err != nil {
			panic(err)
		}

		logrus.AddHook(elkHook)
	}

	if isSyslog {
		var syslogHook *lSyslog.SyslogHook
		var sysProto = conf.GetString("logging.syslog_protocol")

		if sysProto == "" {
			sysProto = "udp"
		}

		if syslogHook, err = lSyslog.NewSyslogHook(sysProto, syslogURL, syslog.LOG_DEBUG, bc.Component); err != nil {
			panic(err)
		}

		logrus.AddHook(syslogHook)
	}

	return Logger{logrus.StandardLogger()}
}
