package tracker

import (
	"backend/internal/logger"
	"backend/internal/storage"
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tongo/wallet"
	"go.uber.org/zap"
)

type Tracker struct {
	ctx                          context.Context
	storage                      storage.Storage
	client                       *tonapi.Client
	wallet                       *wallet.Wallet
	blackTicketCollectionAddress string
	whiteTicketCollectionAddress string
	raffleAddress                string
}

type Func[T any] func() (T, error)

func infinityRateLimitRetry[T any](
	fn Func[T],
) (T, error) {
	for {
		result, err := fn()
		if err != nil {
			var e *tonapi.ErrorStatusCode
			if errors.As(errors.Unwrap(err), &e) && e.StatusCode == 429 {
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}

		return result, err
	}
}

func NewTracker(ctx context.Context) *Tracker {

	if err := godotenv.Load(); err != nil {
		panic("no .env file found")
	}

	walletMnemonic := os.Getenv("WALLET_MNEMONIC")
	walletVersion := os.Getenv("WALLET_VERSION")
	logger.Debug("tracker initialization: .env provided data", zap.String("wallet version", walletVersion), zap.Bool("wallet mnemonic", walletMnemonic != ""))

	sqliteStorage := storage.NewSqliteStorage()

	logger.Debug("tracker initialization: tonapi client...\n")
	client, err := tonapi.NewClient(tonapi.TonApiURL, &tonapi.Security{})
	if err != nil {
		panic("tracker initialization: failed to initialize tonapi client")
	}

	logger.Debug("tracker initialization:  wallet...\n")
	pk, err := wallet.SeedToPrivateKey(walletMnemonic)
	if err != nil {
		panic("tracker initialization: failed to initialize wallet")
	}

	logger.Debug("tracker initialization:", zap.Bool("private key is not empty", pk != nil))
	version := WalletMap[walletVersion]

	logger.Debug("tracker initialization: wallet info", zap.String("version", walletVersion), zap.Int("version index", version), zap.Bool("private key is empty", pk == nil))
	oracleWallet, err := wallet.New(pk, wallet.Version(version), client)

	if err != nil {
		panic("failed to initialize wallet, possible wrong mnemonic")
	}

	logger.Debug("tracker initialization: initializing tracker... done")
	return &Tracker{
		ctx:                          ctx,
		storage:                      sqliteStorage,
		client:                       client,
		wallet:                       &oracleWallet,
		raffleAddress:                os.Getenv("RAFFLE_ADDRESS"),
		blackTicketCollectionAddress: os.Getenv("BLACK_TICKET_COLLECTION_ADDRESS"),
		whiteTicketCollectionAddress: os.Getenv("WHITE_TICKET_COLLECTION_ADDRESS"),
	}
}

func (t *Tracker) GetRaffleAccountDeployedLt() (int64, error) {
	var lastTraceID *tonapi.TraceID = nil

	raffleAccountID, err := ton.ParseAccountID(t.raffleAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to parse raffle address", zap.String("raffle address", t.raffleAddress), zap.Error(err))
		return 0, err
	}

	var beforeLt = tonapi.OptInt64{Value: 0, Set: false}
	for {
		if lastTraceID != nil {
			lastTrace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: lastTraceID.GetID()})
				},
			)
			if err != nil {
				return 0, err
			}
			beforeLt = tonapi.NewOptInt64(lastTrace.Transaction.Lt)
		}

		logger.Debug("verify raffle account: search first traceID")
		accountTracesResult, err := infinityRateLimitRetry(
			func() (*tonapi.TraceIDs, error) {
				return t.client.GetAccountTraces(t.ctx, tonapi.GetAccountTracesParams{
					AccountID: raffleAccountID.ToRaw(),
					Limit:     tonapi.NewOptInt(GlobalLimitWindowSize),
					BeforeLt:  beforeLt,
				})
			})

		if err != nil {
			logger.Fatal("verify raffle account: failed to search first traceID", zap.Error(err))
			return 0, err
		}

		if len(accountTracesResult.Traces) > 0 {
			lastTraceID = &accountTracesResult.Traces[len(accountTracesResult.Traces)-1]
		}

		logger.Debug("verify raffle account: check conditions in proper to continue", zap.Int("trace count", len(accountTracesResult.Traces)))
		if len(accountTracesResult.Traces) < GlobalLimitWindowSize {
			break
		}
	}

	if lastTraceID == nil {
		return 0, errors.New("verify raffle account: no traces found")
	}

	lastTrace, err := infinityRateLimitRetry(
		func() (*tonapi.Trace, error) {
			return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: lastTraceID.GetID()})
		},
	)

	if err != nil {
		return 0, err
	}

	return lastTrace.Transaction.GetLt(), nil
}

func (t *Tracker) Run(raffleStartedLt int64) {

	log.Printf("\n\n GATHERING CANDIDATE REGISTRATIONS \n\n")

	err := t.collectCandidateRegistrationActions(t.raffleAddress, raffleStartedLt)
	if err != nil {
		panic("failed to collect candidate registration actions")
	}

	log.Printf("\n\n GATHERING WHITE TICKET MINTED \n\n")
	err = t.collectWhiteTicketMintedActions(raffleStartedLt)
	if err != nil {
		panic("failed to collect white ticket mints")
	}

	log.Printf("\n\n GATHERING BLACK TICKET PURCHASES \n\n")
	err = t.collectBlackTicketPurchasedActions(raffleStartedLt)
	if err != nil {
		panic("failed to collect black ticket purchase")
	}

	log.Printf("\n\n GATHERING PARTICIPANT REGISTRATIONS \n\n")
	err = t.collectParticipantRegistrationActions(t.raffleAddress, raffleStartedLt)
	if err != nil {
		panic("failed to collect participant registration actions")
	}

	log.Printf("\n\n BLOCKCHAIN SYNCHRONIZATION \n\n")
	err = t.synchronize()
	if err != nil {
		panic("failed to synchronize blockchain")
	}
}

func (t *Tracker) Finalize() {
	log.Printf("Tracker stopped.\n")
}
