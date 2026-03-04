package agent

import (
	"context"

	"go.uber.org/zap"
)

func Example() {
	reportInterval := uint(10)
	pollInterval := uint(2)
	signKey := "secret"

	log, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initiate logger")
	}

	agent := New("localhost:8080", reportInterval, pollInterval, signKey, nil, nil, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = agent.StartGathering(ctx)
	if err != nil {
		log.Fatal("failed to launch agent", zap.Error(err))
	}
}
