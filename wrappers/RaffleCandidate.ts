import {
    Address,
    beginCell,
    Cell,
    Contract,
    contractAddress,
    ContractProvider
} from '@ton/core';

export type RaffleCandidateConfiguration = {raffleAddress: Address; userAddress: Address};

export function raffleCandidateConfigurationToCell(config: RaffleCandidateConfiguration): Cell {
    return beginCell()
        .storeAddress(config.raffleAddress)
        .storeAddress(config.userAddress)
        .endCell();
}

export class RaffleCandidate implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {

        return new RaffleCandidate(address);
    }

    static createFromConfig(config: RaffleCandidateConfiguration, code: Cell, workchain = 0) {
        const data = raffleCandidateConfigurationToCell(config);
        const init = { code, data };
        return new RaffleCandidate(contractAddress(workchain, init), init);
    }

    async getUserAddress(provider: ContractProvider) {
        const result = await provider.get('userAddress', []);
        return result.stack.readAddress();
    }

    async getParticipantIndex(provider: ContractProvider) {
        const result = await provider.get('participantIndex', []);
        return result.stack.readNumberOpt();
    }
}
