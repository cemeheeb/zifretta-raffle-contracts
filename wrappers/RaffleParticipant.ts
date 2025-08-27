import { Address, beginCell, Cell, Contract, contractAddress, ContractProvider, Sender, SendMode } from '@ton/core';

export class RaffleParticipant implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {

        return new RaffleParticipant(address);
    }

    async getUserAddress(provider: ContractProvider) {
        const result = await provider.get('userAddress', []);
        return result.stack.readAddress();
    }
}
