tolk 1.0

import "storage"
import "errors"
import "messages"
import "constants"

type AllowedMessageCandidate = RaffleCandidateInitialize | RaffleCandidateSetConditions | RaffleCandidateSetParticipantIndex;

fun onInternalMessage(inMessage: InMessage) {

    var storage = lazy RaffleCandidateStorage.load();
    assert (inMessage.senderAddress == storage.raffleAddress) throw ERROR_NOT_AUTHORIZED;

    val inMessageBody = lazy AllowedMessageCandidate.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleCandidateInitialize => {
            assert (inMessageBody.telegramID > 0) throw ERROR_INVALID_TELEGRAM_ID
            assert (storage.telegramID == 0) throw ERROR_ALREADY_REGISTERED
            storage.telegramID = inMessageBody.telegramID;
            storage.save();

            val excessMessage = createMessage({
                bounce: false,
                dest: inMessageBody.recipientAddress,
                value: 0,
                body: createEmptyCell(),
            });

            reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
            excessMessage.send(SEND_MODE_CARRY_ALL_BALANCE | SEND_MODE_IGNORE_ERRORS);
        }
        RaffleCandidateSetConditions => {
            assert (storage.isMatched == false) throw ERROR_ALREADY_MATCHED;

            storage.conditions = inMessageBody.conditions;
            storage.isMatched = inMessageBody.isMatched;
            storage.save();

            if (inMessageBody.isMatched) {
                val approveMessage = createMessage({
                    bounce: true,
                    dest: storage.raffleAddress,
                    value: 0,
                    body: RaffleApprove {
                        userAddress: storage.userAddress,
                    },
                });

                reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
                approveMessage.send(SEND_MODE_CARRY_ALL_BALANCE);
            } else {
                val excessMessage = createMessage({
                    bounce: false,
                    dest: inMessageBody.recipientAddress,
                    value: 0,
                    body: createEmptyCell(),
                });

                reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_CANDIDATE_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
                excessMessage.send(SEND_MODE_CARRY_ALL_BALANCE | SEND_MODE_IGNORE_ERRORS);
            }
        }
        RaffleCandidateSetParticipantIndex => {
            storage.participantIndex = inMessageBody.participantIndex;
            storage.save();
        }
        else => {
            assert (inMessage.body.isEmpty()) throw ERROR_WRONG_OPCODE;
        }
    }
}

fun onBouncedMessage(in: InMessageBounced) {
    in.bouncedBody.skipBouncedPrefix();

    // process only raffle approve bounces; on other messages, an exception will be thrown, it's okay
    RaffleApprove.fromSlice(in.bouncedBody);

    var storage = lazy RaffleCandidateStorage.load();
    storage.isMatched = false;
    storage.save();
}

get fun conditions(): bits256 {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.conditions;
}

get fun participantIndex(): int? {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.participantIndex;
}
