
type ForwardPayloadRemainder = RemainingBitsAndRefs

struct (0x13370010) RaffleRegisterCandidate {
}

struct (0x13370011) RaffleSetConditions {
    userAddress: address;
    conditions: bits256;
}

struct (0x13370012) RaffleApprove {
    userAddress: address;
}

struct (0x13370020) RaffleCandidateSetConditions {
    recipientAddress: address;
    conditions: bits256;
    isMatched: bool;
}

struct (0x13370021) RaffleCandidateSetParticipantIndex {
    participantIndex: uint64;
}

struct (0x13370030) RaffleParticipantSetUserAddress {
    userAddress: address;
}