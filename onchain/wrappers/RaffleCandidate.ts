import {
    Address,
    Cell,
    Contract,
    ContractProvider
} from '@ton/core';
import {Conditions} from "./types";

export type RaffleCandidateData = {
    conditions: Conditions;
    telegramID: number | null;
    participantIndex: number | null;
}

export class RaffleCandidate implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {

        return new RaffleCandidate(address);
    }

    async getData(provider: ContractProvider): Promise<RaffleCandidateData> {
        const result = await provider.get('raffleCandidateData', []);

        const conditionsSlice = result.stack.readCell().beginParse();
        return {
            conditions: {
                whiteTicketMinted: conditionsSlice.loadUint(8),
                blackTicketPurchased: conditionsSlice.loadUint(8),
            },
            telegramID: result.stack.readNumberOpt(),
            participantIndex: result.stack.readNumberOpt(),
        };
    }
}
