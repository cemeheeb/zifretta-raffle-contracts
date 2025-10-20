import {createContext} from "react";

import {BlockchainData} from "@workers";

export type BlockchainContextData = BlockchainData & {
  sendCandidateRegistration: () => Promise<void>;
}

export const BlockchainContext = createContext<BlockchainData>({
  raffles: []
});

export const BlockchainContextProvider = BlockchainContext.Provider;
