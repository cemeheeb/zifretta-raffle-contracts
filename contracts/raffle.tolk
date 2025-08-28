tolk 1.0

import "@stdlib/common"
import "constants"
import "errors"
import "storage"
import "messages"

type AllowedMessageRaffleParticipant = RaffleRegisterCandidate | RaffleSetConditions | RaffleApprove;

fun onInternalMessage(inMessage: InMessage) {
    val inMessageBody = lazy AllowedMessageRaffleParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleRegisterCandidate => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_CANDIDATE_STORAGE + REGISTER_CANDIDATE_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();

            assert (blockchain.now() < storage.deadline) throw ERROR_EXPIRED;

            val deployRaffleCandidateMessage = createMessage({
                bounce: false,
                dest: calculateRaffleCandidateAddress(contract.getAddress(), inMessage.senderAddress, storage.candidateCode),
                value: 0,
                body: createEmptyCell()
            });

            deployRaffleCandidateMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }
        RaffleSetConditions => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_PARTICIPANT_STORAGE + APPROVE_CANDIDATE_OPERATION_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();
            assert (inMessage.senderAddress == storage.ownerAddress) throw ERROR_NOT_AUTHORIZED;

            val candidate = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            val candidateAddress = address.fromValidBuilder(candidate.buildAddress());

            val setConditionsMessage = createMessage({
                bounce: true,
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
            assert (inMessage.valueCoins >= MIN_TONS_FOR_PARTICIPANT_STORAGE + APPROVE_CANDIDATE_OPERATION_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;
            var storage = lazy RaffleStorage.load();

            val raffleCandidate = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            val raffleCandidateAddress = raffleCandidate.buildAddress();
            assert (inMessage.senderAddress == address.fromValidBuilder(raffleCandidateAddress)) throw ERROR_NOT_AUTHORIZED;

            val deployParticipantMessage = createMessage({
                bounce: false,
                dest: calculateRaffleParticipantAddress(contract.getAddress(), storage.participantQuantity, storage.participantCode),
                value: MIN_TONS_FOR_PARTICIPANT_STORAGE + APPROVE_CANDIDATE_OPERATION_FEE,
                body: RaffleParticipantSetUserAddress {
                    userAddress: inMessageBody.userAddress,
                },
            });
            deployParticipantMessage.send(SEND_MODE_PAY_FEES_SEPARATELY);

            val setParticipantIndex = createMessage({
                bounce: true,
                dest: raffleCandidateAddress,
                value: APPROVE_CANDIDATE_OPERATION_FEE,
                body: RaffleCandidateSetParticipantIndex {
                    participantIndex: storage.participantQuantity
                },
            });
            setParticipantIndex.send(SEND_MODE_PAY_FEES_SEPARATELY);

            storage.participantQuantity += 1;
            storage.save();
        }
        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}

get fun staticData(): (int, bits256, int) {
    val storage = lazy RaffleStorage.load();
    return (storage.deadline, storage.conditions, storage.participantQuantity);
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
