tolk 1.0

import "storage"
import "errors"
import "messages"
import "constants"

type AllowedMessageCandidate = RaffleCandidateInitialize | RaffleCandidateSetConditions | RaffleCandidateSetParticipantIndex;

fun onInternalMessage(inMessage: InMessage) {
    reserveToncoinsOnBalance(MIN_TONS_FOR_CANDIDATE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);

    var storage = lazy RaffleCandidateStorage.load();

    assert (inMessage.senderAddress == storage.raffleAddress) throw ERROR_NOT_AUTHORIZED;

    val inMessageBody = lazy AllowedMessageCandidate.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleCandidateInitialize => {
            storage.telegramID = inMessageBody.telegramID;
            storage.save();
        }
        RaffleCandidateSetConditions => {
            assert (storage.isMatched == false) throw ERROR_ALREADY_MATCHED;

            storage.conditions = inMessageBody.conditions;
            storage.isMatched = inMessageBody.isMatched;

            if (inMessageBody.isMatched) {
                val approveMessage = createMessage({
                    bounce: true,
                    dest: storage.raffleAddress,
                    value: 0,
                    body: RaffleApprove {
                        userAddress: storage.userAddress,
                    },
                });

                approveMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
            } else {
                val excessMessage = createMessage({
                    bounce: true,
                    dest: inMessageBody.recipientAddress,
                    value: 0,
                    body: createEmptyCell(),
                });

                excessMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
            }

            storage.save();
        }
        RaffleCandidateSetParticipantIndex => {
            storage.participantIndex = inMessageBody.participantIndex;
            storage.save();
        }
        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}

get fun conditions(): bits256 {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.conditions;
}

get fun participantIndex(): int? {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.participantIndex;
}
