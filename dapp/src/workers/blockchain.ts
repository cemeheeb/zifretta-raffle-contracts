import {Address, TupleItemInt, TupleItemSlice} from "@ton/core";
import {TonApiClient, TraceIDs} from "@ton-api/client";

import {BlockchainData, RaffleData, RaffleCandidateData, RaffleParticipantData} from "./types";
import {infinityRetry} from "./utils";

interface BlockchainDataRequest {
  oracleAddress: string;
  userAddress: string;
}

interface BlockchainDataResponse {
  data: BlockchainData
}

// eslint-disable-next-line no-restricted-globals
const context: Worker = self as any;
const client: TonApiClient = new TonApiClient();

const getBlockchainRaffleData = async (raffleAddress: Address): Promise<RaffleData> => {
  const raffleData = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleAddress, "raffleData")
  );

  return {
    address: raffleAddress,
    minCandidateQuantity: (raffleData.stack[0] as TupleItemInt).value,
    conditionsDuration: (raffleData.stack[1] as TupleItemInt).value,
    conditions: {
      blackTicketPurchased: BigInt((raffleData.stack[2] as TupleItemSlice).cell.asSlice().loadUint(8)),
      whiteTicketMinted: BigInt((raffleData.stack[2] as TupleItemSlice).cell.asSlice().loadUint(8))
    },
    minCandidateReachedLt: raffleData.stack.length > 3 ? (raffleData.stack[3] as TupleItemInt).value : 0n,
    minCandidateReachedUnixTime: raffleData.stack.length > 4 ? (raffleData.stack[4] as TupleItemInt).value : 0n,
    candidatesQuantity: raffleData.stack.length > 5 ? (raffleData.stack[5] as TupleItemInt).value : 0n,
    participantsQuantity: raffleData.stack.length > 6 ? (raffleData.stack[6] as TupleItemInt).value : 0n,
    winnersQuantity: raffleData.stack.length > 7 ? (raffleData.stack[7] as TupleItemInt).value : 0n,
    winners: [],
  }
}

const getBlockchainRaffleCandidateData = async (raffleCandidateAddress: Address): Promise<RaffleCandidateData> => {
  const raffleCandidateData = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleCandidateAddress, "raffleCandidateData")
  );

  return {
    address: raffleCandidateAddress,
    conditions: {
      blackTicketPurchased: BigInt((raffleCandidateData.stack[0] as TupleItemSlice).cell.asSlice().loadUint(8)),
      whiteTicketMinted: BigInt((raffleCandidateData.stack[0] as TupleItemSlice).cell.asSlice().loadUint(8))
    },
    participantIndex: (raffleCandidateData.stack[1] as TupleItemInt).value,
  }
}

const getBlockchainRaffleParticipantData = async (raffleParticipantAddress: Address): Promise<RaffleParticipantData> => {
  const raffleCandidateData = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleParticipantAddress, "raffleParticipantData")
  );

  return {
    address: raffleParticipantAddress,
    participantIndex: (raffleCandidateData.stack[0] as TupleItemInt).value,
    userAddress: raffleCandidateData.stack.length > 1 ? (raffleCandidateData.stack[1] as TupleItemSlice).cell.beginParse().loadAddress() : null,
    winnerIndex: raffleCandidateData.stack.length > 2 ? (raffleCandidateData.stack[2] as TupleItemInt).value : null
  }
}


const getBlockchainRaffleCandidateAddress = async (raffleAddress: Address, userAddress: Address): Promise<Address> => {
  const raffleCandidateAddressResult = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleAddress, "raffleCandidateAddress")
  );

  return (raffleCandidateAddressResult.stack.pop() as TupleItemSlice).cell.beginParse().loadAddress();
}

const getBlockchainRaffleParticipantAddress = async (raffleAddress: Address, participantIndex: bigint): Promise<Address> => {
  const raffleParticipantAddressResult = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleAddress, "raffleParticipantAddress")
  );

  return (raffleParticipantAddressResult.stack.pop() as TupleItemSlice).cell.beginParse().loadAddress();
}

const getBlockchainData = async (oracleAddress: Address, userAddress: Address): Promise<BlockchainDataResponse> => {

  const traceFrame = 100;

  let blockchainAccountTracesResult: TraceIDs;
  let beforeLt: bigint;

  do {
    blockchainAccountTracesResult = await infinityRetry(() => client.accounts.getAccountTraces(oracleAddress));
  } while (blockchainAccountTracesResult.traces.length < traceFrame);

  const blockchainData: BlockchainData = {
    raffles: []
  };

  for (const traceID of blockchainAccountTracesResult.traces) {

    const traceResult = await infinityRetry(() => client.traces.getTrace(traceID.id))
    for (const message of traceResult.transaction.outMsgs.filter(message => !message.bounced)) {

      if (message.opCode === BigInt(0x33010000) && message.destination) {
        const raffleData = await getBlockchainRaffleData(message.destination?.address);
        const raffleCandidateAddress = await getBlockchainRaffleCandidateAddress(message.destination?.address, userAddress);
        const raffleCandidateData = await getBlockchainRaffleCandidateData(raffleCandidateAddress);

        if (raffleCandidateData.participantIndex !== null) {

          const raffleParticipantAddress = await getBlockchainRaffleParticipantAddress(message.destination?.address, raffleCandidateData.participantIndex);
          const raffleParticipantData = await getBlockchainRaffleParticipantData(raffleParticipantAddress);

          blockchainData.raffles.push({
            raffleData,
            raffleCandidateData,
            raffleParticipantData,
            winnersData: []
          });
        }
      }
    }
  }

  return {data: blockchainData};
};

context.onmessage = async (e: MessageEvent<BlockchainDataRequest>) => {
  try {
    const {oracleAddress, userAddress} = e.data;
    const blockchainDataResponse = await getBlockchainData(Address.parse(oracleAddress), Address.parse(userAddress));
    context.postMessage({type: 'success', data: blockchainDataResponse});
  } catch (error) {
    context.postMessage({
      type: 'error',
      error: error instanceof Error ? error.message : 'Unknown error'
    });
  }
};

// Для TypeScript
export {};
