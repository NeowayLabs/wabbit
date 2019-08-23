package utils

import (
	"testing"

	"github.com/NeowayLabs/wabbit"
	"github.com/streadway/amqp"
)

func TestConvertOptDefaults(t *testing.T) {
	opt, err := ConvertOpt(nil)

	if err != nil {
		t.Error(err)
		return
	}

	if opt.ContentType != "text/plain" {
		t.Errorf("Invalid opt content type: %s", opt.ContentType)
	}

	if opt.ContentEncoding != "" {
		t.Errorf("Invalid opt encoding: %s", opt.ContentEncoding)
	}

	if opt.DeliveryMode != amqp.Transient {
		t.Errorf("Invalid default delivery mode: %d\n", opt.DeliveryMode)
	}

	if opt.Priority != uint8(0) {
		t.Errorf("Invalid default priority: %d\n", opt.Priority)
	}

	if opt.MessageId != "" {
		t.Errorf("Invalid default message ID: %s\n", opt.MessageId)
	}

	if len(opt.Headers) != 0 {
		t.Errorf("Invalid value for headers: %v", opt.Headers)
	}

	if opt.CorrelationId != "" {
		t.Errorf("Invalid correlation ID: %s\n", opt.CorrelationId)
	}

	if opt.ReplyTo != "" {
		t.Errorf("Invalid ReplyTo: %s\n", opt.ReplyTo)
	}
}

func TestConvertOpt(t *testing.T) {
	opt, err := ConvertOpt(wabbit.Option{
		"contentType": "binary/fuzz",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.ContentType != "binary/fuzz" {
		t.Errorf("Wrong value for content type: %s", opt.ContentType)
	}

	opt, err = ConvertOpt(wabbit.Option{
		"contentEncoding": "bleh",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.ContentEncoding != "bleh" {
		t.Errorf("Invalid value for contentEncoding: %s", opt.ContentEncoding)
	}

	opt, err = ConvertOpt(wabbit.Option{
		"contentEncoding": "bleh",
		"contentType":     "binary/fuzz",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.ContentType != "binary/fuzz" {
		t.Errorf("Wrong value for content type: %s", opt.ContentType)
	}

	if opt.ContentEncoding != "bleh" {
		t.Errorf("Invalid value for contentEncoding: %s", opt.ContentEncoding)
	}

	opt, err = ConvertOpt(wabbit.Option{
		"messageId": "12345",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.MessageId != "12345" {
		t.Errorf("Invalid value for messageId: %s", opt.MessageId)
	}

	opt, err = ConvertOpt(wabbit.Option{
		"correlationId": "0123-4414",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.CorrelationId != "0123-4414" {
		t.Errorf("Invalid value for correlationId: %s", opt.CorrelationId)
	}

	opt, err = ConvertOpt(wabbit.Option{
		"replyTo": "rpc-queue",
	})

	if err != nil {
		t.Error(err)
		return
	}

	if opt.ReplyTo != "rpc-queue" {
		t.Errorf("Invalid value for ReplyTo: %s", opt.ReplyTo)
	}

	// setting invalid value

	opt, err = ConvertOpt(wabbit.Option{
		"NotExists": "bleh",
	})

	if err == nil {
		t.Errorf("Shall fail...")
		return
	}
}
