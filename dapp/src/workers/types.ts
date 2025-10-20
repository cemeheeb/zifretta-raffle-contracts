export type RaffleConditions = {
  blackTicketPurchased: bigint;
  whiteTicketMinted: bigint;
}

export type RaffleData = {
  address: string;
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
  address: string;
  conditions: RaffleConditions;
  participantIndex?: bigint;
}

export type RaffleParticipantData = {
  address: string;
  participantIndex: bigint;
  userAddress?: string;
  winnerIndex?: bigint;
}

export type Raffle = {
  raffleData: RaffleData;
  raffleCandidateData?: RaffleCandidateData;
  raffleParticipantData?: RaffleParticipantData;
  winnersData: RaffleParticipantData[]
}

export type BlockchainData = {
  raffles: Raffle[];
}

export enum RaffleState {
  Qualification,
  Waiting,
  Conditions,
  Timer,
  Participation,
  Result
}

export enum RaffleProgressStep {
  Waiting,
  Timer,
  Participation,
}


