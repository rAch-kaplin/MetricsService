package ping

import (
	"context"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Usecase interface {
	Check(ctx context.Context) error
}

type PingUsecase struct {
	pinger Pinger
}

func NewPingUsecase(pinger Pinger) *PingUsecase {
	return &PingUsecase{pinger: pinger}
}

func (puc *PingUsecase) Check(ctx context.Context) error {
	return puc.pinger.Ping(ctx)
}
