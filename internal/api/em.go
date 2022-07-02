package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	apipb "github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/em"
	"google.golang.org/grpc"
)

func init() {
	reg = append(reg, &emApi{})
}

type emApi struct {
	apipb.UnimplementedEmServiceServer
	BaseHandler
}

func (ema *emApi) Desc() *grpc.ServiceDesc {
	return &apipb.EmService_ServiceDesc
}

func (ema *emApi) Send(ctx context.Context, req *apipb.EmSendRequest) (*apipb.EmSendResponse, error) {
	fromsk, err := ema.a.n.ID().Find(req.From)
	if err != nil {
		return nil, errors.Wrap(err, "finding identity for sender")
	}

	from, err := fromsk.PublicIdentity()
	if err != nil {
		return nil, errors.Wrap(err, "building public from sender identity")
	}

	tmpl := &em.Email{
		From:    from,
		Time:    time.Now().Unix(),
		Headers: req.Headers,
		Parts:   make([]em.EmailPart, 0, len(req.Parts)),
	}

	for _, p := range req.Parts {
		tmpl.Parts = append(tmpl.Parts, em.EmailPart{
			Mime: p.Mime,
			Data: p.Data,
		})
	}

	var wg sync.WaitGroup
	errs := []error{}
	var mu sync.Mutex

	for _, to := range req.To {
		wg.Add(1)
		go func(to string, tmpl *em.Email) {
			defer func() {
				if r := recover(); r != nil {
					logging.Entry().WithField("err", fmt.Sprintf("%s", r)).Error("recovered from panic")
					errs = append(errs, errors.New("internal error sending to "+to))
				}
			}()

			defer wg.Done()
			em := tmpl.Copy()

			id, err := ema.a.n.Resolver().Find(to)
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

func (ema *emApi) sendSingle(ctx context.Context, em *em.Email) error {
	return ema.a.n.Em().Send(ctx, em)
}
