import "utils"

type ForwardPayloadRemainder = RemainingBitsAndRefs

struct (0x13370010) RaffleRegisterCandidate {
    telegramID: uint64;
}

struct (0x13370011) RaffleSetConditions {
    userAddress: address;
    conditions: bits256;
}

struct (0x13370012) RaffleApprove {
    userAddress: address;
}

struct (0x13370013) RaffleNext {
    message: Cell<SnakeString>;
}

struct (0x13370014) RafflePayout {
}

struct (0x13370020) RaffleCandidateInitialize {
    telegramID: uint64;
    recipientAddress: address;
}

struct (0x13370021) RaffleCandidateSetConditions {
    recipientAddress: address;
    conditions: bits256;
    isMatched: bool;
}

struct (0x13370022) RaffleCandidateSetParticipantIndex {
    participantIndex: uint64;
}

struct (0x13370030) RaffleParticipantSetUserAddress {
    userAddress: address;
}

struct (0x13370031) RaffleParticipantRewardNotification {
    message: Cell<SnakeString>
}