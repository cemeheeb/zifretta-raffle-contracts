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
import {OperationCodes} from "./constants";

type RaffleConditions = {
  blackTicketPurchased: bigint;
  whiteTicketMinted: bigint;
}

export type RaffleConfiguration = {
    ownerAddress: Address;
    minParticipantsQuantity: bigint;
    conditions: RaffleConditions;
    duration: bigint;
};

export type RaffleData = {
    minParticipantsQuantity: bigint;
    conditions: RaffleConditions;
    duration: bigint;
    minCandidateReachedLt: bigint;
    minCandidateReachedUnixTime: bigint;
    candidatesQuantity: bigint;
    participantsQuantity: bigint;
    winnersQuantity: bigint;
    winners: number[];
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
      .storeUint(configuration.minParticipantsQuantity, 32)
      .storeBits(raffleConditionConfigurationToBits256(configuration.conditions))
      .storeUint(configuration.duration, 32)
      .storeRef(raffleCandidateCode)
      .storeRef(raffleParticipantCode)
      .storeUint(0, 64)
      .storeInt(0, 64)
      .storeUint(0, 64)
      .storeUint(0, 64)
      .storeUint(0, 8)
      .storeMaybeRef(null)
      .endCell();
}

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
            recipientAddress: Address;
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
            forwardAmount: bigint;
            message: string;
        }
    ) {

        await provider.internal(via, {
            value: options.value,
            sendMode: SendMode.PAY_GAS_SEPARATELY,
            body: beginCell()
                .storeUint(OperationCodes.OP_RAFFLE_NEXT, 32)
                .storeCoins(options.forwardAmount)
                .storeStringTail(options.message)
                .endCell(),
        });
    }

    async getData(provider: ContractProvider): Promise<RaffleData> {
        const result = await provider.get('raffleData', []);
        const minCandidateQuantity = BigInt(result.stack.readNumber());
        const participantDuration = BigInt(result.stack.readNumber());

        const conditionBuffer = result.stack.readBuffer();
        const conditionSlice = new Slice(new BitReader(new BitString(conditionBuffer, 0, conditionBuffer.length)), []);
        const conditions = {
            blackTicketPurchased: BigInt(conditionSlice.loadUint(8)),
            whiteTicketMinted: BigInt(conditionSlice.loadUint(8)),
        }

        const minCandidateReachedLt = BigInt(result.stack.readNumber());
        const minCandidateReachedUnixTime = BigInt(result.stack.readNumber());
        const candidatesQuantity = BigInt(result.stack.readNumber());
        const participantsQuantity = BigInt(result.stack.readNumber());
        const winnersQuantity = BigInt(result.stack.readNumber());

        return {
            minParticipantsQuantity: minCandidateQuantity,
            duration: participantDuration,
            conditions,
            minCandidateReachedLt,
            minCandidateReachedUnixTime,
            candidatesQuantity,
            participantsQuantity,
            winnersQuantity,
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
