import {
    Address,
    beginCell, BitReader, BitString,
    Cell,
    Contract,
    contractAddress,
    ContractProvider,
    Sender,
    SendMode,
    Slice
} from '@ton/core';
import {compile} from "@ton/blueprint";

type RaffleConditions = {
  blackTicketPurchased: bigint;
  whiteTicketMinted: bigint;
}

export type RaffleConfiguration = {
    ownerAddress: Address;
    deadline: bigint;
    maxRewards: bigint;
    conditions: RaffleConditions;
};

export type RaffleStaticData = {
    deadline: bigint;
    maxRewards: bigint;
    conditions: RaffleConditions;
    participantsQuantity: bigint;
    nextRewardIndex: bigint;
    winners: bigint[];
}

function raffleConditionConfigurationToBits256(configuration: RaffleConditions) {

  return beginCell()
      .storeUint(configuration.blackTicketPurchased, 8)
      .storeUint(configuration.whiteTicketMinted, 8)
      .storeUint(0, 240) // fill zeroes remaining bits
      .asSlice()
      .loadBits(256);
}

export async function raffleConfigurationToCell(configuration: RaffleConfiguration): Promise<Cell> {

  const raffleCandidateCode = await compile("RaffleCandidate");
  const raffleParticipantCode = await compile("RaffleParticipant");

  return beginCell()
      .storeAddress(configuration.ownerAddress)
      .storeUint(configuration.deadline, 64)
      .storeUint(configuration.maxRewards, 32)
      .storeBits(raffleConditionConfigurationToBits256(configuration.conditions))
      .storeRef(raffleCandidateCode)
      .storeRef(raffleParticipantCode)
      .storeUint(0, 64)
      .storeUint(0, 32)
      .storeMaybeRef(null)
      .endCell();
}

export const OperationCodes = {
    OP_RAFFLE_REGISTER_CANDIDATE: 0x13370010,
    OP_RAFFLE_SET_CONDITIONS: 0x13370011,
    OP_RAFFLE_APPROVE: 0x13370012,
    OP_RAFFLE_NEXT: 0x13370013,
    OP_RAFFLE_CANDIDATE_INITIALIZE: 0x13370020,
    OP_RAFFLE_CANDIDATE_SET_CONDITIONS: 0x13370021,
    OP_RAFFLE_CANDIDATE_SET_PARTICIPANT_INDEX: 0x13370022,
    OP_RAFFLE_PARTICIPANT_SET_USER_ADDRESS: 0x13370030,
    OP_RAFFLE_PARTICIPANT_REWARD_NOTIFICATION: 0x13370031,
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
            telegramID: bigint;
        }
    ) {
        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_RAFFLE_REGISTER_CANDIDATE, 32)
                .storeUint(options.telegramID, 64)
                .endCell(),
        });
    }

    async sendConditions(
        provider: ContractProvider,
        via: Sender,
        options: {
            value: bigint;
            userAddress: Address;
            conditions: RaffleConditions
        }
    ) {
        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_RAFFLE_SET_CONDITIONS, 32)
                .storeAddress(options.userAddress)
                .storeBits(raffleConditionConfigurationToBits256(options.conditions))
                .endCell(),
        });
    }

    async sendRaffleNext(
        provider: ContractProvider,
        via: Sender,
        options: {
            value: bigint;
            message: string | null;
        }
    ) {

        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_RAFFLE_NEXT, 32)
                .storeMaybeStringRefTail(options.message)
                .endCell(),
        });
    }

    async getStaticData(provider: ContractProvider): Promise<RaffleStaticData> {
        const result = await provider.get('staticData', []);

        const deadline = BigInt(result.stack.readNumber());
        const maxRewards = BigInt(result.stack.readNumber());
        const conditionBuffer = result.stack.readBuffer();
        const conditionSlice = new Slice(new BitReader(new BitString(conditionBuffer, 0, conditionBuffer.length)), []);
        const conditions = {
            blackTicketPurchased: BigInt(conditionSlice.loadUint(8)),
            whiteTicketMinted: BigInt(conditionSlice.loadUint(8)),
        }
        const participantsQuantity = BigInt(result.stack.readNumber());
        const nextRewardIndex = BigInt(result.stack.readNumber());

        return {
            deadline,
            maxRewards,
            conditions,
            participantsQuantity,
            nextRewardIndex,
            winners: []
        };
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
