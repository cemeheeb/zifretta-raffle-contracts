import {BlockchainData, Raffle, RaffleState} from "./types";

export const selectRaffle = (blockchainData: BlockchainData, raffleAddress: string): Raffle | undefined => {

  return blockchainData.raffles
    .findLast((raffle: Raffle) => raffle.raffleData.address === raffleAddress);
}

export const selectLaunchedRafflesData = (blockchainData: BlockchainData) => {

  return blockchainData.raffles
    .filter((raffle: Raffle) => raffle.raffleData.winnersQuantity === 0n)
    .map((raffle: Raffle) => raffle.raffleData);
}

export const selectCompletedRafflesData = (blockchainData: BlockchainData) => {

  return blockchainData.raffles
    .filter((raffle: Raffle) => raffle.raffleData.winnersQuantity > 0n)
    .map((raffle: Raffle) => raffle.raffleData);
}

const selectRaffleRemainingSeconds = (raffle: Raffle) => {

  return Math.max(
    Number(raffle.raffleData?.minCandidateReachedUnixTime) + Number(raffle.raffleData?.conditionsDuration) - (new Date().getTime() / 1000),
    0
  );
}

export const selectRaffleState = (blockchainData: BlockchainData, raffleAddress: string) => {

  const raffle = blockchainData.raffles.filter((raffle: Raffle) => raffleAddress.toLowerCase() === raffle.raffleData.address.toLowerCase()).pop()
  if (!raffle) {
    return null;
  }

  const candidateIsDeployed = raffle.raffleCandidateData !== undefined;
  const minimalCandidateQuantityReached = raffle.raffleData.candidatesQuantity >= raffle.raffleData.minCandidateQuantity;
  const conditionsReached = raffle.raffleCandidateData?.conditions?.blackTicketPurchased === raffle.raffleData?.conditions?.blackTicketPurchased
    && raffle.raffleCandidateData?.conditions?.whiteTicketMinted === raffle.raffleData?.conditions?.whiteTicketMinted;
  const raffleRemainingSeconds = selectRaffleRemainingSeconds(raffle);

  if (!candidateIsDeployed) {
    return RaffleState.Qualification;
  }

  if (!minimalCandidateQuantityReached) {
    return RaffleState.Waiting;
  }

  if (!conditionsReached) {
    return RaffleState.Conditions;
  }

  if (raffleRemainingSeconds > 0) {
    return RaffleState.Timer;
  }

  if (raffle.raffleData.candidatesQuantity < raffle.raffleData.minCandidateQuantity) {
    return RaffleState.Participation;
  }

  return RaffleState.Result;
}
