package grpc

import (
	"fmt"

	"github.com/hellobchain/memmq/broker"
	"github.com/hellobchain/memmq/core/log"
	mq "github.com/hellobchain/memmq/proto"
	"github.com/hellobchain/wswlog/wlogging"
	"golang.org/x/net/context"
)

var logger = wlogging.MustGetFileLoggerWithoutName(log.LogConfig)

type handler struct{}

func (h *handler) Pub(ctx context.Context, req *mq.PubRequest) (*mq.PubResponse, error) {
	logger.Infof("Pub topic: %s payload: %s", req.Topic, string(req.Payload))
	if err := broker.Publish(req.Topic, req.Payload); err != nil {
		return nil, fmt.Errorf("pub error: %v", err)
	}
	return new(mq.PubResponse), nil
}

func (h *handler) Sub(req *mq.SubRequest, stream mq.MQ_SubServer) error {
	logger.Infof("Sub topic: %s", req.Topic)
	ch, err := broker.Subscribe(req.Topic)
	if err != nil {
		return fmt.Errorf("could not subscribe: %v", err)
	}
	defer broker.Unsubscribe(req.Topic, ch)

	for p := range ch {
		logger.Infof("Sub payload: %s", string(p))
		if err := stream.Send(&mq.SubResponse{Payload: p}); err != nil {
			return fmt.Errorf("failed to send payload: %v", err)
		}
	}

	return nil
}
