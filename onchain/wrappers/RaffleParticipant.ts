import { Address, Cell, Contract, ContractProvider } from '@ton/core';

export type RaffleParticipantData = {
    participantIndex: number;
    userAddress: Address | null;
    winnerIndex: number | null;
}

export class RaffleParticipant implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {

        return new RaffleParticipant(address);
    }

    async getData(provider: ContractProvider): Promise<RaffleParticipantData> {
        const result = await provider.get('raffleParticipantData', []);
        return {
            participantIndex: result.stack.readNumber(),
            userAddress: result.stack.readAddress(),
            winnerIndex: result.stack.readNumber(),
        };
    }
}
