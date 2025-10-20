import {Address} from '@ton/core';

export type RaffleConditions = {
  blackTicketPurchased: bigint;
  whiteTicketMinted: bigint;
}

export type RaffleData = {
  address: Address;
  minCandidateQuantity: bigint;
  conditionsDuration: bigint;
  conditions: RaffleConditions;
  minCandidateReachedLt: bigint;
  minCandidateReachedUnixTime: bigint;
  candidatesQuantity: bigint;
  participantsQuantity: bigint;
  winnersQuantity: bigint;
  winners: string[];
}

export type RaffleCandidateData = {
  address: Address;
  conditions: RaffleConditions;
  participantIndex: bigint | null;
}

export type RaffleParticipantData = {
  address: Address;
  participantIndex: bigint;
  userAddress: Address | null;
  winnerIndex: bigint | null;
}

export type Raffle = {
  raffleData: RaffleData;
  raffleCandidateData: RaffleCandidateData;
  raffleParticipantData: RaffleParticipantData;
  winnersData: RaffleParticipantData[]
}

export type BlockchainData = {
  raffles: Raffle[];
}

export const selectorActiveRafflesData = (blockchainData: BlockchainData) => {
  return blockchainData.raffles
    .filter((raffle: Raffle) => raffle.raffleData.minCandidateReachedUnixTime > Date.now() / 1000)
    .map((raffle: Raffle) => raffle.raffleData);
}

export const selectorNonActiveRafflesData = (blockchainData: BlockchainData) => {
  return blockchainData.raffles
    .filter((raffle: Raffle) => raffle.raffleData.minCandidateReachedUnixTime <= Date.now() / 1000)
    .map((raffle: Raffle) => raffle.raffleData);
}


