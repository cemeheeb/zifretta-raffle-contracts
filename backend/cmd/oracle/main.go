package main

import (
	"backend/internal/logger"
	"backend/internal/tracker"
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для ошибок
	errCh := make(chan error, 1)

	// Запускаем горутину с основным циклом
	go func() {
		logger.Initialize(logger.Configuration{
			LogFile: "tracker.log",
			Level:   zapcore.InfoLevel,
			Console: true,
		})
		trackerInstance := tracker.NewTracker(ctx)

		raffleAccountData, err := trackerInstance.GetRaffleAccountData()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case <-ctx.Done():
				trackerInstance.Finalize()
				return
			default:
				trackerInstance.Run(
					61946738000007, // ХАРДКОД LT от 26 сентября приблизительно с 20:00 по Мск
					raffleAccountData.Conditions.WhiteTicketMinted,
					raffleAccountData.Conditions.BlackTicketPurchased,
				)
			}
		}
	}()

	// Ожидаем ошибку или сигнал завершения
	select {
	case err := <-errCh:
		logger.Fatal("fatal error", zap.Error(err))
		cancel()
	case <-waitForInterrupt():
		logger.Info("gracefully shutting down...")
		cancel()
	}
}

func waitForInterrupt() <-chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	return sigCh
}
