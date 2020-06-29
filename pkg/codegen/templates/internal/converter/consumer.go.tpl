package converter

import (
	"context"
	"os"

	stan "github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/discovery"
)

// ConsumerService is the template for a job processing service.
type ConsumerService struct{}

// Run starts the consumerservice
func (c *ConsumerService) Run(ctx context.Context, log logrus.FieldLogger) error {
	///StartBlock(channels)
	{{- if .channels }}
{{ .channels }}
	{{- else }}
	channels := map[string]stan.MsgHandler{}
	{{- end }}
	///EndBlock(channels)

	endpoint, err := discovery.Find("nats")
	if err != nil {
		return errors.Wrap(err, "failed to find nats")
	}

	clientID, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "failed to get hostname")
	}

	// TODO(jaredallard): handle connection loss
	sc, err := stan.Connect("test-cluster", clientID, stan.NatsURL(endpoint))
	if err != nil {
		return errors.Wrap(err, "failed to create nats client")
	}

	subs := make([]stan.Subscription, 0)
	for queue, c := range channels {
		// TODO(jaredallard): we should handle this at the queue level
		sub, err := sc.QueueSubscribe(queue, queue, c,
			stan.MaxInflight(1), stan.SetManualAckMode(), stan.DurableName(queue),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create subscription for queue '%s'", queue)
		}

		subs = append(subs, sub)
	}

	log.Infof("created consumers")
	// wait for the context to cancel
	<-ctx.Done()

	log.Infof("shutting down")

	// we can technically use defer in the for loop, but this is more explicit
	for _, s := range subs {
		s.Close()
	}

	// shutdown the connection
	sc.Close()

	return nil
}
