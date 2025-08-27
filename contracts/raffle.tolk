tolk 1.0

import "@stdlib/common"
import "constants"
import "errors"
import "storage"
import "messages"

type AllowedMessageRaffleParticipant = AskToRegisterCandidate | AskToApproveCandidate | ApproveCandidate;

fun onInternalMessage(inMessage: InMessage) {
    val inMessageBody = lazy AllowedMessageRaffleParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        AskToRegisterCandidate => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_CANDIDATE_STORAGE + REGISTER_CANDIDATE_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();

            assert (blockchain.now() < storage.deadline) throw ERROR_EXPIRED;

            val deployRaffleCandidateMessage = createMessage({
                bounce: false,
                dest: calculateRaffleCandidateAddress(contract.getAddress(), inMessage.senderAddress, storage.candidateCode),
                value: 0,
                body: beginCell().endCell()
            });

            deployRaffleCandidateMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }
        AskToApproveCandidate => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_PARTICIPANT_STORAGE + APPROVE_CANDIDATE_OPERATION_FEE * 3) throw ERROR_INSUFFICIENT_FUNDS;

            var storage = lazy RaffleStorage.load();
            assert (inMessage.senderAddress == storage.ownerAddress) throw ERROR_NOT_AUTHORIZED;

            val destination = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            val registerParticipantMessage = createMessage({
                bounce: true,
                dest: address.fromValidBuilder(destination.buildAddress()),
                value: 0,
                body: ApproveCandidate {
                    userAddress: inMessageBody.userAddress,
                    participantIndex: storage.participantQuantity
                },
            });

            registerParticipantMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
            storage.participantQuantity += 1;

            storage.save();
        }
        ApproveCandidate => {
            assert (inMessage.valueCoins >= MIN_TONS_FOR_PARTICIPANT_STORAGE + APPROVE_CANDIDATE_OPERATION_FEE) throw ERROR_INSUFFICIENT_FUNDS;

            val storage = lazy RaffleStorage.load();

            val raffleCandidateAddress = calculateRaffleCandidateAddress(contract.getAddress(), inMessageBody.userAddress, storage.candidateCode);
            assert (inMessage.senderAddress == address.fromValidBuilder(raffleCandidateAddress.buildAddress())) throw ERROR_NOT_AUTHORIZED;

            val registerParticipantMessage = createMessage({
                bounce: false,
                dest: calculateRaffleParticipantAddress(contract.getAddress(), inMessageBody.participantIndex, storage.participantCode),
                value: inMessage.valueCoins,
                body: inMessageBody,
            });
            registerParticipantMessage.send(SEND_MODE_PAY_FEES_SEPARATELY);
        }
        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}

get fun deadline(): int {
    val storage = lazy RaffleStorage.load();
    return storage.deadline;
}

get fun participantQuantity(): int {
    val storage = lazy RaffleStorage.load();
    return storage.participantQuantity;
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
