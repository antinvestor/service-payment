package events

import (
	"context"
	"errors"
	"github.com/pitabwire/frame"
)

type JengaGoodsServices struct {
	Service *frame.Service
}

func (event *JengaGoodsServices) Name() string {
	return "jenga.goods.services"
}

func (event *JengaGoodsServices) PayloadType() any {
	pType := ""
	return &pType
}

func (event *JengaGoodsServices) Validate(ctx context.Context, payload any) error {
	if _, ok := payload.(*string); !ok {
		return errors.New(" payload is not of type string")
	}

	return nil
}

func (event *JengaGoodsServices) Execute(ctx context.Context, payload any) error {
	paymentID := *payload.(*string)
	logger := event.Service.L(ctx).WithField("payload", paymentID).WithField("type", event.Name())
	logger.Debug("handling event")

}
