tolk 1.0

import "storage"
import "messages"
import "constants"

type AllowedMessageParticipant = ApproveCandidate

fun onInternalMessage(inMessage: InMessage) {
    reserveToncoinsOnBalance(MIN_TONS_FOR_PARTICIPANT_STORAGE, RESERVE_MODE_EXACT_AMOUNT | RESERVE_MODE_BOUNCE_ON_ACTION_FAIL);

    var storage = lazy RaffleParticipantStorage.load();
    assert (storage.raffleAddress == inMessage.senderAddress) throw 0x2000;

    val inMessageBody = lazy AllowedMessageParticipant.fromSlice(inMessage.body);

    match (inMessageBody) {
        ApproveCandidate => {
            storage.userAddress = inMessageBody.userAddress;
            storage.save()
        }

        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}

get fun userAddress(): address? {
    val storage = lazy RaffleParticipantStorage.load();
    return storage.userAddress;
}
