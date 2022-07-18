package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	apipb "github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/comm"
	"google.golang.org/grpc"
)

func init() {
	reg = append(reg, &commApi{})
}

type commApi struct {
	apipb.UnimplementedEmServiceServer
	BaseHandler
}

func (ema *commApi) Desc() *grpc.ServiceDesc {
	return &apipb.EmService_ServiceDesc
}

func (ema *commApi) Send(ctx context.Context, req *apipb.EmSendRequest) (*apipb.EmSendResponse, error) {
	fromsk, err := ema.a.n.ID().Find(req.From)
	if err != nil {
		return nil, errors.Wrap(err, "finding identity for sender")
	}

	from, err := fromsk.PublicIdentity()
	if err != nil {
		return nil, errors.Wrap(err, "building public from sender identity")
	}

	tmpl := &comm.Message{
		From:        from,
		Time:        time.Now().Unix(),
		Body:        req.Headers,
		Attachments: make([]comm.MessageAttachment, 0, len(req.Parts)),
	}

	for _, p := range req.Parts {
		tmpl.Attachments = append(tmpl.Attachments, comm.MessageAttachment{
			Mime: p.Mime,
			Data: p.Data,
		})
	}

	var wg sync.WaitGroup
	errs := []error{}
	var mu sync.Mutex

	for _, to := range req.To {
		wg.Add(1)
		go func(to string, tmpl *comm.Message) {
			defer func() {
				if r := recover(); r != nil {
					logging.Entry().WithField("err", fmt.Sprintf("%s", r)).Error("recovered from panic")
					errs = append(errs, errors.New("internal error sending to "+to))
				}
			}()

			defer wg.Done()
			em := tmpl.Copy()

			id, err := ema.a.n.Resolver().ResolvePI(to)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = append(errs, errors.Wrap(err, "resolving recipient"))
				return
			}

			em.To = id

			if err = ema.sendSingle(ctx, em); err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = append(errs, errors.Wrap(err, "sending to "+to))
				return
			}
		}(to, tmpl)
	}

	wg.Wait()

	r := &apipb.EmSendResponse{}
	if len(errs) > 0 {
		r.Errors = make([]string, 0, len(errs))
	}
	for _, err := range errs {
		r.Errors = append(r.Errors, err.Error())
	}

	return r, nil
}

func (ema *commApi) sendSingle(ctx context.Context, em *comm.Message) error {
	return ema.a.n.Comm().Send(ctx, em)
}
