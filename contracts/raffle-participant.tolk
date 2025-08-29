tolk 1.0

import "storage"
import "messages"
import "constants"
import "errors"

type AllowedMessageParticipant = RaffleParticipantSetUserAddress | RaffleParticipantRewardNotification;

fun onInternalMessage(inMessage: InMessage) {
    reserveToncoinsOnBalance(MIN_TONS_FOR_PARTICIPANT_STORAGE, RESERVE_MODE_EXACT_AMOUNT | RESERVE_MODE_BOUNCE_ON_ACTION_FAIL);

    var storage = lazy RaffleParticipantStorage.load();
    assert (storage.raffleAddress == inMessage.senderAddress) throw ERROR_NOT_AUTHORIZED;

    val inMessageBody = lazy AllowedMessageParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        RaffleParticipantSetUserAddress => {
            storage.userAddress = inMessageBody.userAddress;
            storage.save();
        }

        RaffleParticipantRewardNotification => {
            assert (storage.userAddress != null) throw ERROR_INVALID_PARTICIPANT;
            val excessMessage = createMessage({
                bounce: true,
                dest: storage.userAddress,
                value: 0,
                body: inMessageBody.message,
            });

            excessMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw ERROR_WRONG_OPCODE
        }
    }
}

get fun userAddress(): address? {
    val storage = lazy RaffleParticipantStorage.load();
    return storage.userAddress;
}
