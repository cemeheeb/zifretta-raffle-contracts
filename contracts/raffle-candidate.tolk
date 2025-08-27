tolk 1.0

import "storage"
import "messages"
import "constants"

type AllowedMessageCandidate = ApproveCandidate;

fun onInternalMessage(inMessage: InMessage) {

    var storage = lazy RaffleCandidateStorage.load();

    assert (inMessage.senderAddress == storage.raffleAddress) throw 0x2000;

    val inMessageBody = lazy AllowedMessageCandidate.fromSlice(inMessage.body);

    match (inMessageBody) {
        ApproveCandidate => {
            var storage = lazy RaffleCandidateStorage.load();

            val registerParticipantMessage = createMessage({
                bounce: true,
                dest: inMessage.senderAddress,
                value: 0,
                body: inMessageBody,
            });

            registerParticipantMessage.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
            storage.participantIndex = inMessageBody.participantIndex;

            storage.save();
        }
        else => {
            // ignore empty messages, "wrong opcode" for others
            assert (inMessage.body.isEmpty()) throw 0xFFFF
        }
    }
}


get fun userAddress(): address {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.userAddress;;
}

get fun participantIndex(): int? {
    val storage = lazy RaffleCandidateStorage.load();
    return storage.participantIndex;
}
