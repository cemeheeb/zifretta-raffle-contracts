package tracker

import (
	"backend/internal/logger"

	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
)

func (t *Tracker) VerifyRaffleAccount() error {
	logger.Debug("verify raffle account: verifying raffle address...")

	raffleAccountID, err := ton.ParseAccountID(t.raffleAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to parse raffle address", zap.String("raffle address", t.raffleAddress), zap.Error(err))
		return err
	}

	logger.Debug("verify raffle account: validating raffle address", zap.String("raffle address", t.raffleAddress))

	raffleAccount, err := infinityRateLimitRetry(
		func() (*tonapi.Account, error) {
			return t.client.GetAccount(t.ctx, tonapi.GetAccountParams{
				AccountID: t.raffleAddress,
			})
		})

	if err != nil {
		logger.Fatal("verify raffle account: failed to get raffle account state")
		return err
	}

	logger.Debug("verify raffle account: raffle contract info:", zap.Int64("balance", raffleAccount.GetBalance()))

	raffleData, err := infinityRateLimitRetry(
		func() (*tonapi.MethodExecutionResult, error) {
			return t.client.ExecGetMethodForBlockchainAccount(t.ctx, tonapi.ExecGetMethodForBlockchainAccountParams{
				AccountID:  t.raffleAddress,
				MethodName: "raffleData",
				Args:       make([]string, 0),
			})
		})

	if err != nil {
		logger.Fatal("verify raffle account: failed to get raffleData, invalid raffle contract")
		return err
	}

	logger.Debug("verify raffle account: raffleData", zap.Bool("success", raffleData.GetSuccess()))

	raffleCandidateAddressResult, err := infinityRateLimitRetry(
		func() (*tonapi.MethodExecutionResult, error) {
			return t.client.ExecGetMethodWithBodyForBlockchainAccount(t.ctx,
				tonapi.OptExecGetMethodWithBodyForBlockchainAccountReq{
					Value: tonapi.ExecGetMethodWithBodyForBlockchainAccountReq{
						Args: []tonapi.ExecGetMethodArg{
							{Value: raffleAccountID.ToRaw(), Type: "slice"},
						},
					},
					Set: true,
				},
				tonapi.ExecGetMethodWithBodyForBlockchainAccountParams{
					AccountID:  t.raffleAddress,
					MethodName: "raffleCandidateAddress",
				},
			)
		})

	if err != nil {
		logger.Fatal("verify raffle account: failed to get raffle candidate address, invalid raffle contract")
		return err
	}

	raffleCandidateAccountAddressSliceOpt := raffleCandidateAddressResult.GetStack()[0].GetCell()
	cell, err := boc.DeserializeBocHex(raffleCandidateAccountAddressSliceOpt.Value)
	if err != nil {
		logger.Fatal("verify raffle account: failed to deserialize raffle candidate address, invalid raffle contract")
		return err
	}

	var raffleCandidateAccountAddress tlb.MsgAddress
	err = tlb.Unmarshal(cell[0], &raffleCandidateAccountAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to extract raffle candidate address from boc, invalid raffle contract")
		return err
	}

	raffleCandidateAccountID, err := tongo.AccountIDFromTlb(raffleCandidateAccountAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to read raffle candidate address due to address tlb scheme, invalid raffle contract")
		return err
	}

	logger.Debug("verify raffle account:", zap.String("raffle candidate address", raffleCandidateAccountID.ToHuman(true, false)))

	raffleParticipantAddressResult, err := infinityRateLimitRetry(
		func() (*tonapi.MethodExecutionResult, error) {
			return t.client.ExecGetMethodWithBodyForBlockchainAccount(t.ctx,
				tonapi.OptExecGetMethodWithBodyForBlockchainAccountReq{
					Value: tonapi.ExecGetMethodWithBodyForBlockchainAccountReq{
						Args: []tonapi.ExecGetMethodArg{
							{Value: "1", Type: "tinyint"},
						},
					},
					Set: true,
				},
				tonapi.ExecGetMethodWithBodyForBlockchainAccountParams{
					AccountID:  t.raffleAddress,
					MethodName: "raffleParticipantAddress",
				},
			)
		})

	if err != nil {
		logger.Fatal("verify raffle account: failed to get raffle participant address from raffle contract, invalid raffle contract")
		return err
	}

	raffleParticipantAccountAddressSliceOpt := raffleParticipantAddressResult.GetStack()[0].GetCell()
	cell, err = boc.DeserializeBocHex(raffleParticipantAccountAddressSliceOpt.Value)
	if err != nil {
		logger.Fatal("verify raffle account: failed to extract raffle participant address from boc, invalid raffle contract")
		return err
	}

	var raffleParticipantAccountAddress tlb.MsgAddress
	err = tlb.Unmarshal(cell[0], &raffleParticipantAccountAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to read raffle participant address due to address tlb scheme, invalid raffle contract")
		return err
	}

	raffleParticipantAccountID, err := tongo.AccountIDFromTlb(raffleParticipantAccountAddress)
	if err != nil {
		logger.Fatal("verify raffle account: invalid raffle participant address, invalid raffle contract")
		return err
	}
	logger.Debug("raffleParticipant account", zap.String("address", raffleParticipantAccountID.ToHuman(true, false)))
	logger.Debug("Verifying raffle address... done")
	return nil
}
