import "constants"
import "@stdlib/common"

struct RaffleStorage {
    // state init:
    ownerAddress: address;
    deadline: uint64;
    conditions: bits256;
    candidateCode: cell;
    participantCode: cell;
    // after initialization state
    participantQuantity: uint64; // default: 0
}

fun RaffleStorage.load() {
    return RaffleStorage.fromCell(contract.getData());
}

fun RaffleStorage.save(self) {
    contract.setData(self.toCell());
}

struct RaffleCandidateStorage {
    // state init:
    raffleAddress: address;
    userAddress: address;
    // after initialization state
    conditions: bits256;
    isMatched: bool;
    participantIndex: uint64?;
}

fun RaffleCandidateStorage.load() {
    return RaffleCandidateStorage.fromCell(contract.getData())
}

fun RaffleCandidateStorage.save(self) {
    contract.setData(self.toCell())
}

struct RaffleParticipantStorage {
    // state init:
    raffleAddress: address;
    participantIndex: uint64;
    // after initialization state
    userAddress: address?;
}

fun RaffleParticipantStorage.load() {
    return RaffleParticipantStorage.fromCell(contract.getData())
}

fun RaffleParticipantStorage.save(self) {
    contract.setData(self.toCell())
}

@pure
fun createBits256None(): bits256
    asm "x{0000000000000000000000000000000000000000000000000000000000000000} PUSHSLICE"

fun calculateRaffleCandidateAddress(raffleAddress: address, userAddress: address, code: cell): AutoDeployAddress {
    val data: RaffleCandidateStorage = { raffleAddress, userAddress, conditions: createBits256None(), isMatched: false, participantIndex: null };

    return {
        workchain: MY_WORKCHAIN,
        stateInit: {
            code,
            data: data.toCell()
        }
    }
}

fun calculateRaffleParticipantAddress(raffleAddress: address, participantIndex: uint64, code: cell): AutoDeployAddress {
    val data: RaffleParticipantStorage = { raffleAddress, participantIndex, userAddress: createAddressNone() };

    return {
        workchain: MY_WORKCHAIN,
        stateInit: {
            code,
            data: data.toCell()
        }
    }
}