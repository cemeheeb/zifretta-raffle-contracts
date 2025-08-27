
type ForwardPayloadRemainder = RemainingBitsAndRefs

struct (0x70000000) AskToRegisterCandidate {
}

struct (0x70000001) RegisterCandidate {
    userAddress: address;
}

struct (0x80000000) AskToApproveCandidate {
    userAddress: address;
}

struct (0x80000001) ApproveCandidate {
    userAddress: address;
    participantIndex: uint64;
}
