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
	"github.com/tonkeeper/tongo/liteapi"
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
	client, err := tonapi.NewClient(tonapi.TonApiURL, tonapi.WithToken("AF64UYO7BZZBSYIAAAAGMH67OZFW62PFAP6HGNCLST5YRXESM6FBPBYPEVZDGI3RDCSEUYY"))
	if err != nil {
		panic(err)
	}

	logger.Debug("tracker initialization:  wallet...\n")
	clientLite, err := liteapi.NewClientWithDefaultMainnet()
	if err != nil {
		panic(err)
	}

	pk, err := wallet.SeedToPrivateKey(walletMnemonic)
	if err != nil {
		panic(err)
	}

	logger.Debug("tracker initialization:", zap.Bool("private key is not empty", pk != nil))
	version := WalletMap[walletVersion]

	logger.Debug("tracker initialization: wallet info", zap.String("version", walletVersion), zap.Int("version index", version), zap.Bool("private key is empty", pk == nil))
	oracleWallet, err := wallet.New(pk, wallet.Version(version), clientLite)

	if err != nil {
		panic(err)
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

func (t *Tracker) Run(raffleDeployedLt int64, minCandidateReachedLt int64, maxParticipantUnixTime int64) {

	log.Printf("\n\n GATHERING CANDIDATE REGISTRATIONS \n\n")

	err := t.collectCandidateRegistrationActions(t.raffleAddress, raffleDeployedLt)
	if err != nil {
		panic(err)
	}

	log.Printf("\n\n GATHERING WHITE TICKET MINTED \n\n")
	err = t.collectWhiteTicketMintedActions(raffleDeployedLt)
	if err != nil {
		panic(err)
	}

	log.Printf("\n\n GATHERING BLACK TICKET PURCHASES \n\n")
	err = t.collectBlackTicketPurchasedActions(raffleDeployedLt)
	if err != nil {
		panic(err)
	}

	log.Printf("\n\n GATHERING PARTICIPANT REGISTRATIONS \n\n")
	err = t.collectParticipantRegistrationActions(t.raffleAddress, raffleDeployedLt)
	if err != nil {
		panic(err)
	}

	log.Printf("\n\n BLOCKCHAIN SYNCHRONIZATION \n\n")
	err = t.synchronize()
	if err != nil {
		panic(err)
	}
}

func (t *Tracker) Finalize() {
	log.Printf("Tracker stopped.\n")
}
