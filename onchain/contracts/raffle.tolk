tolk 1.0

import "@stdlib/tvm-dicts"
import "@stdlib/common"
import "@stdlib/gas-payments"
import "constants"
import "errors"
import "storage"
import "messages"

type AllowedMessageRaffleParticipant = RaffleRegisterCandidate | RaffleSetConditions | RaffleApprove | RaffleNext;

fun onInternalMessage(inMessage: InMessage) {
    val inMessageBody = lazy AllowedMessageRaffleParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleRegisterCandidate => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE + RAFFLE_REGISTER_CANDIDATE_OPERATION_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();

            assert (blockchain.now() < storage.deadline) throw ERROR_EXPIRED;

            val deployRaffleCandidateMessage = createMessage({
                bounce: false,
                dest: calculateRaffleCandidateAddress(contract.getAddress(), inMessage.senderAddress, storage.candidateCode),
                value: 0,
                body: RaffleCandidateInitialize {
                    telegramID: inMessageBody.telegramID,
                    recipientAddress: inMessage.senderAddress
                }
            });

            reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
            deployRaffleCandidateMessage.send(SEND_MODE_CARRY_ALL_BALANCE);
        }
        RaffleSetConditions => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE + RAFFLE_SET_CONDITIONS_OPERATION_FEE) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();
            assert (inMessage.senderAddress == storage.ownerAddress) throw ERROR_NOT_AUTHORIZED;

            val candidate = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            val candidateAddress = address.fromValidBuilder(candidate.buildAddress());

            val setConditionsMessage = createMessage({
                bounce: false,
                dest: candidateAddress,
                value: 0,
                body: RaffleCandidateSetConditions {
                    recipientAddress: inMessage.senderAddress,
                    conditions: inMessageBody.conditions,
                    isMatched: (storage.conditions as slice).bitsEqual(inMessageBody.conditions as slice)
                },
            });

            setConditionsMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }
        RaffleApprove => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE + MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE + RAFFLE_APPROVE_OPERATION_FEE + RAFFLE_PARTICIPANT_SET_USER_ADDRESS_OPERATION_FEE) throw ERROR_INSUFFICIENT_FUNDS;
            var storage = lazy RaffleStorage.load();

            val raffleCandidate = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            val raffleCandidateAddress = address.fromValidBuilder(raffleCandidate.buildAddress());
            assert (inMessage.senderAddress == raffleCandidateAddress) throw ERROR_NOT_AUTHORIZED;

            val deployParticipantMessage = createMessage({
                bounce: false,
                dest: calculateRaffleParticipantAddress(contract.getAddress(), storage.participantQuantity, storage.participantCode),
                value: MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE + RAFFLE_PARTICIPANT_SET_USER_ADDRESS_OPERATION_FEE,
                body: RaffleParticipantSetUserAddress {
                    userAddress: inMessageBody.userAddress,
                },
            });
            deployParticipantMessage.send(SEND_MODE_PAY_FEES_SEPARATELY);

            val setParticipantIndexMessage = createMessage({
                bounce: false,
                dest: raffleCandidateAddress,
                value: MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE + RAFFLE_CANDIDATE_SET_PARTICIPANT_INDEX_OPERATION_FEE,
                body: RaffleCandidateSetParticipantIndex {
                    participantIndex: storage.participantQuantity
                },
            });
            setParticipantIndexMessage.send(SEND_MODE_PAY_FEES_SEPARATELY);

            val excessMessage = createMessage({
                bounce: false,
                dest: storage.ownerAddress,
                value: 0,
                body: createEmptyCell(),
            });

            storage.participantQuantity += 1;
            storage.save();

            reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
            excessMessage.send(SEND_MODE_CARRY_ALL_BALANCE);
        }
        RaffleNext => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE + RAFFLE_NEXT_OPERATION_FEE) throw ERROR_INSUFFICIENT_FUNDS;
            var storage = lazy RaffleStorage.load();

            assert (storage.nextRewardIndex + 1 < storage.maxRewards) throw ERROR_REWARD_LIMIT_REACHED;

            // random.initialize();

            var rewardIndex: uint64 = 2;
            do {
                rewardIndex = random.range(storage.participantQuantity);

                var success: bool = false;
                if (storage.winners != null) {
                    (_, success) = storage.winners.uDictGetRef(64, rewardIndex);
                }
            } while (success);

            storage.winners.uDictSetRef(64, rewardIndex, createEmptyCell());
            storage.nextRewardIndex += 1;
            storage.save();

            val raffleParticipantAddress = calculateRaffleParticipantAddress(contract.getAddress(), rewardIndex, storage.participantCode);
            val participantRewardMessage = createMessage({
                bounce: false,
                dest: address.fromValidBuilder(raffleParticipantAddress.buildAddress()),
                value: 0,
                body: RaffleParticipantRewardNotification {
                    message: inMessageBody.message
                },
            });

            participantRewardMessage.send(SEND_MODE_CARRY_ALL_BALANCE | SEND_MODE_IGNORE_ERRORS);
        }
        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}

get fun staticData(): (int, int, bits256, int, int, dict) {
    val storage = lazy RaffleStorage.load();
    return (storage.deadline, storage.maxRewards, storage.conditions, storage.participantQuantity, storage.nextRewardIndex, storage.winners);
}

get fun raffleCandidateAddress(userAddress: address): address {
    val storage = lazy RaffleStorage.load();
    val candidateAddress = calculateRaffleCandidateAddress(contract.getAddress(), userAddress, storage.candidateCode);
    return address.fromValidBuilder(candidateAddress.buildAddress());
}

get fun raffleParticipantAddress(participantIndex: uint64): address {
    val storage = lazy RaffleStorage.load();
    val participantAddress = calculateRaffleParticipantAddress(contract.getAddress(), participantIndex, storage.participantCode);
    return address.fromValidBuilder(participantAddress.buildAddress());
}
