import {
    Address,
    Cell,
    Contract,
    ContractProvider
} from '@ton/core';


export class RaffleCandidate implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {

        return new RaffleCandidate(address);
    }

    async getConditions(provider: ContractProvider) {
        const result = await provider.get('conditions', []);
        return result.stack.readBuffer();
    }

    async getParticipantIndex(provider: ContractProvider) {
        const result = await provider.get('participantIndex', []);
        return result.stack.readNumberOpt();
    }
}
