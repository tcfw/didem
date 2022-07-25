package consensus

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	dataTopic = "did_data"

	pubsubBuf = 10
)

type p2p struct {
	router *pubsub.PubSub
	logger *logrus.Entry

	topics map[string]*pubsub.Topic
}

func (p *p2p) Msgs(channel string) (<-chan *Msg, error) {
	var err error

	t, ok := p.topics[channel]
	if !ok {
		t, err = p.router.Join(channel)
		if err != nil {
			return nil, errors.Wrap(err, "join data topic")
		}

		p.topics[channel] = t
	}

	sub, err := t.Subscribe()
	if err != nil {
		return nil, errors.Wrap(err, "subscribing to data topic")
	}

	msgCh := make(chan *Msg, pubsubBuf)

	go func() {
		for {
			m, err := sub.Next(context.Background())
			if err != nil {
				p.logger.WithError(err).Errorf("sub %s closed", channel)
				close(msgCh)
				return
			}

			msg := &Msg{}
			if err := msgpack.Unmarshal(m.Data, msg); err != nil {
				p.logger.WithError(err).WithField("from", m.From).Error("unmarshalling msg")
				continue
			}
			msg.Signature = m.Signature
			msg.Key = m.Key

			msgCh <- msg
		}
	}()

	return msgCh, nil
}

func (p *p2p) PublishContext(ctx context.Context, channel string, m *Msg) error {
	t, ok := p.topics[channel]
	if !ok {
		return errors.New("not subscribed")
	}

	b, err := m.Marshal()
	if err != nil {
		return errors.Wrap(err, "")
	}

	return t.Publish(ctx, b, pubsub.WithReadiness(pubsub.MinTopicSize(1)))
}
