tolk 1.0

import "storage"
import "messages"
import "constants"
import "errors"

type AllowedMessageParticipant = RaffleParticipantInitialize | RaffleParticipantWinnerNotification;

fun onInternalMessage(inMessage: InMessage) {

    var storage = lazy RaffleParticipantStorage.load();
    assert (storage.raffleAddress == inMessage.senderAddress) throw ERROR_NOT_AUTHORIZED;

    val inMessageBody = lazy AllowedMessageParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleParticipantInitialize => {
            storage.userAddress = inMessageBody.userAddress;
            storage.save();
        }

        RaffleParticipantWinnerNotification => {
            assert (inMessage.valueCoins > MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE + inMessageBody.forwardAmount) throw ERROR_NOT_ENOUGH_TONS;
            assert (storage.userAddress != null) throw ERROR_INVALID_PARTICIPANT;

            storage.winnerIndex = inMessageBody.winnerIndex;
            storage.save();

            val userNotificationMessage = createMessage({
                bounce: true,
                dest: storage.userAddress,
                value: inMessageBody.forwardAmount,
                body: inMessageBody.forwardPayload,
            });

            reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
            userNotificationMessage.send(SEND_MODE_PAY_FEES_SEPARATELY);

            val excessMessage = createMessage({
                bounce: true,
                dest: storage.raffleAddress,
                value: 0,
                body: createEmptyCell(),
            });

            reserveToncoinsOnBalance(MIN_TONS_FOR_RAFFLE_PARTICIPANT_STORAGE, RESERVE_MODE_EXACT_AMOUNT);
            excessMessage.send(SEND_MODE_CARRY_ALL_BALANCE | SEND_MODE_IGNORE_ERRORS);
        }

        else => {
            assert (inMessage.body.isEmpty()) throw ERROR_WRONG_OPCODE
        }
    }
}

get fun raffleParticipantData(): (int, address?, int?) {
    val storage = lazy RaffleParticipantStorage.load();
    return (storage.participantIndex, storage.userAddress, storage.winnerIndex);
}
