import {Address, TupleItemInt, TupleItemSlice} from "@ton/core";
import {TonApiClient, Transactions} from "@ton-api/client";

import {RAFFLE_BOC_HASH} from "./constants";
import {
  BlockchainData,
  RaffleData,
  RaffleCandidateData,
  RaffleParticipantData
} from "./types";
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

const getBlockchainRaffleData = async (raffleAddress: string): Promise<RaffleData> => {
  const raffleData = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(Address.parse(raffleAddress), "raffleData")
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

  console.log(raffleCandidateData);

  return {
    address: raffleCandidateAddress.toRawString(),
    conditions: {
      blackTicketPurchased: BigInt((raffleCandidateData.stack[0] as TupleItemSlice).cell.asSlice().loadUint(8)),
      whiteTicketMinted: BigInt((raffleCandidateData.stack[0] as TupleItemSlice).cell.asSlice().loadUint(8))
    },
    participantIndex: raffleCandidateData.stack.length > 2 ? (raffleCandidateData.stack[2] as TupleItemInt).value : undefined,
  }
}

const getBlockchainRaffleParticipantData = async (raffleParticipantAddress: Address): Promise<RaffleParticipantData> => {
  const raffleCandidateData = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleParticipantAddress, "raffleParticipantData")
  );

  return {
    address: raffleParticipantAddress.toRawString(),
    participantIndex: (raffleCandidateData.stack[0] as TupleItemInt).value,
    userAddress: raffleCandidateData.stack.length > 1 ? (raffleCandidateData.stack[1] as TupleItemSlice).cell.beginParse().loadAddress().toRawString() : undefined,
    winnerIndex: raffleCandidateData.stack.length > 2 ? (raffleCandidateData.stack[2] as TupleItemInt).value : undefined
  }
}


const getBlockchainRaffleCandidateAddress = async (raffleAddress: Address, userAddress: Address): Promise<Address> => {
  console.log("getBlockchainRaffleCandidateAddress", userAddress.toString({ bounceable: true }));
  const raffleCandidateAddressResult = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleAddress, "raffleCandidateAddress", {
      args: [userAddress.toString({bounceable: true})],
    })
  );

  return (raffleCandidateAddressResult.stack.pop() as TupleItemSlice).cell.beginParse().loadAddress();
}

const getBlockchainRaffleParticipantAddress = async (raffleAddress: Address, participantIndex: bigint): Promise<Address> => {
  console.log("getBlockchainRaffleParticipantAddress", participantIndex);

  const raffleParticipantAddressResult = await infinityRetry(
    () => client.blockchain.execGetMethodForBlockchainAccount(raffleAddress, "raffleParticipantAddress", {
      args: [participantIndex.toString()]
    })
  );

  return (raffleParticipantAddressResult.stack.pop() as TupleItemSlice).cell.beginParse().loadAddress();
}

const getBlockchainData = async (oracleAddress: string, userAddress: string): Promise<BlockchainDataResponse> => {
  const limit = 100;

  let blockchainAccountTransactionsResult: Transactions;
  const blockchainData: BlockchainData = {
    raffles: []
  };

  let beforeLt: bigint | undefined = undefined;

  do {
    blockchainAccountTransactionsResult = await infinityRetry(async () =>
      client.blockchain.getBlockchainAccountTransactions(Address.parse(oracleAddress), {
        before_lt: beforeLt,
        limit
      }));

    for (const transaction of blockchainAccountTransactionsResult.transactions) {
      for (const message of transaction.outMsgs) {
        if (!message) {
          continue;
        }

        if (!message?.init) {
          continue;
        }

        if (message?.init?.boc.beginParse().loadRef().hash().toString("hex") === RAFFLE_BOC_HASH && message?.destination) {
          const raffleData = await getBlockchainRaffleData(message?.destination?.address.toString({ bounceable: true }));
          const raffleCandidateAddress = await getBlockchainRaffleCandidateAddress(message.destination?.address, Address.parse(userAddress));

          let raffleCandidateData: RaffleCandidateData | undefined = undefined;
          try {
            raffleCandidateData = await getBlockchainRaffleCandidateData(raffleCandidateAddress);
            let raffleParticipantData: RaffleParticipantData | undefined = undefined;
            if (raffleCandidateData.participantIndex !== undefined) {
              const raffleParticipantAddress = await getBlockchainRaffleParticipantAddress(message.destination?.address, raffleCandidateData.participantIndex);
              raffleParticipantData = await getBlockchainRaffleParticipantData(raffleParticipantAddress);
            }

            blockchainData.raffles.push({
              raffleData,
              raffleCandidateData,
              raffleParticipantData,
              winnersData: []
            });
          } catch {
            console.log("raffleCandidateData is not deployed");
          }
        }
      }
    }
  } while (blockchainAccountTransactionsResult.transactions.length === limit);

  return {data: blockchainData};
};

context.onmessage = async (e: MessageEvent<BlockchainDataRequest>) => {
  try {
    const {oracleAddress, userAddress} = e.data;
    const blockchainDataResponse = await getBlockchainData(oracleAddress, userAddress);
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
