package main

import (
	"backend/internal/logger"
	"backend/internal/tracker"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для ошибок
	errCh := make(chan error, 1)

	// Запускаем горутину с основным циклом
	go func() {
		logger.Initialize()
		trackerInstance := tracker.NewTracker(ctx)

		err := trackerInstance.VerifyRaffleAccount()
		if err != nil {
			panic(err)
		}

		raffleAccountDeployedAt, err := trackerInstance.GetRaffleAccountDeployedLt()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case <-ctx.Done():
				trackerInstance.Finalize()
				return
			default:
				trackerInstance.Run(raffleAccountDeployedAt)
			}
		}
	}()

	// Ожидаем ошибку или сигнал завершения
	select {
	case err := <-errCh:
		fmt.Printf("Остановка из-за ошибки: %v\n", err)
		cancel()
	case <-waitForInterrupt():
		fmt.Println("Получен сигнал прерывания")
		cancel()
	}
}

func waitForInterrupt() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}
