import { Address, beginCell, Cell, Contract, contractAddress, ContractProvider, Sender, SendMode } from '@ton/core';
import {compile} from "@ton/blueprint";

export type RaffleConfiguration = {
    ownerAddress: Address;
    deadline: bigint;
};

export async function raffleConfigurationToCell(configuration: RaffleConfiguration): Promise<Cell> {

    const raffleCandidateCode = await compile("RaffleCandidate");
    const raffleParticipantCode = await compile("RaffleParticipant");

    return beginCell()
        .storeAddress(configuration.ownerAddress)
        .storeUint(configuration.deadline, 64)
        .storeRef(raffleCandidateCode)
        .storeRef(raffleParticipantCode)
        .storeUint(0, 64)
        .endCell();
}

export const OperationCodes = {
    OP_ASK_TO_REGISTER_CANDIDATE: 0x70000000,
    OP_ASK_TO_APPROVE_CANDIDATE: 0x80000000,
};

export class Raffle implements Contract {
    constructor(readonly address: Address, readonly init?: { code: Cell; data: Cell }) {}

    static createFromAddress(address: Address) {
        return new Raffle(address);
    }

    static async createFromConfig(config: RaffleConfiguration, code: Cell, workchain = 0) {
        const data = await raffleConfigurationToCell(config);
        const init = { code, data };

        return new Raffle(contractAddress(workchain, init), init);
    }

    async sendDeploy(provider: ContractProvider, via: Sender, value: bigint) {
        await provider.internal(via, {
            value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell().endCell(),
        });
    }

    async sendRegisterCandidate(
        provider: ContractProvider,
        via: Sender,
        options: {
            value: bigint;
        }
    ) {
        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_ASK_TO_REGISTER_CANDIDATE, 32)
                .endCell(),
        });
    }

    async sendApproveCandidate(
        provider: ContractProvider,
        via: Sender,
        options: {
            value: bigint;
            userAddress: Address;
        }
    ) {
        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_ASK_TO_APPROVE_CANDIDATE, 32)
                .storeAddress(options.userAddress)
                .endCell(),
        });
    }

    async getDeadline(provider: ContractProvider) {
        const result = await provider.get('deadline', []);
        return result.stack.readNumber();
    }

    async getParticipantQuantity(provider: ContractProvider) {
        const result = await provider.get('participantQuantity', []);
        return result.stack.readNumber();
    }

    async getRaffleCandidateAddress(provider: ContractProvider, address: Address) {
        const result = await provider.get('raffleCandidateAddress', [{ type: 'slice', cell: beginCell().storeAddress(address).endCell() }]);
        return result.stack.readAddress();
    }

    async getRaffleParticipantAddress(provider: ContractProvider, participantIndex: number) {
        const result = await provider.get('raffleParticipantAddress', [{ type: 'int', value: BigInt(participantIndex) }]);
        return result.stack.readAddress();
    }
}
