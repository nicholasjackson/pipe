package config

import (
	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/pipe/logger"
	"github.com/nicholasjackson/pipe/pipe"
	"github.com/nicholasjackson/pipe/providers"
)

func testGetLogger() *logger.LoggerMock {
	return &logger.LoggerMock{
		GetLoggerFunc: func() hclog.Logger {
			return hclog.Default()
		},
		GetStatsDFunc: func() *statsd.Client {
			s, _ := statsd.New("")
			return s
		},
		ProviderConnectionCreatedFunc: func(in1 providers.Provider) {
		},
		ProviderConnectionFailedFunc: func(in1 providers.Provider, in2 error) {
		},
		ProviderMessagePublishedFunc: func(in1 providers.Provider, in2 providers.Message, in3 ...interface{}) {
		},
		ProviderSubcriptionCreatedFunc: func(in1 providers.Provider) {
		},
		ProviderSubcriptionFailedFunc: func(in1 providers.Provider, in2 error) {
		},
		ServerActionPublishFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerActionPublishFailedFunc: func(in1 *pipe.Pipe, in2 providers.Message, in3 error) {
		},
		ServerActionPublishSuccessFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerFailPublishFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerFailPublishFailedFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message, in4 error) {
		},
		ServerFailPublishSuccessFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerHandleMessageExpiredFunc: func(in1 *pipe.Pipe, in2 providers.Message) {
		},
		ServerNewMessageReceivedStartFunc: func(in1 *pipe.Pipe, in2 providers.Message) *logger.LoggerTiming {
			return &logger.LoggerTiming{}
		},
		ServerNoPipesConfiguredFunc: func(in1 providers.Provider) {
		},
		ServerSuccessPublishFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerSuccessPublishFailedFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message, in4 error) {
		},
		ServerSuccessPublishSuccessFunc: func(in1 *pipe.Pipe, in2 *pipe.Action, in3 providers.Message) {
		},
		ServerTemplateProcessFailFunc: func(in1 *pipe.Action, in2 []byte, in3 error) {
		},
		ServerTemplateProcessStartFunc: func(in1 *pipe.Action, in2 []byte) *logger.LoggerTiming {
			return &logger.LoggerTiming{}
		},
		ServerTemplateProcessSuccessFunc: func(in1 *pipe.Action, in2 []byte) {
		},
		ServerUnableToListenFunc: func(in1 providers.Provider, in2 error) {
		},
	}

}
