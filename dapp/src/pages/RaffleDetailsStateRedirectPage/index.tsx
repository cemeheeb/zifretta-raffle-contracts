import {useContext} from "react";
import {useParams, Navigate} from 'react-router-dom';
import {BlockchainContext} from "@contexts";
import {selectRaffleState, RaffleState} from "@workers";

export default () => {
  const { raffleAddress } = useParams();
  const blockchainData = useContext(BlockchainContext);

  switch (selectRaffleState(blockchainData, raffleAddress)) {
    case RaffleState.Qualification:
      return <Navigate to={`/raffles/${raffleAddress}/qualification`}></Navigate>;
    case RaffleState.Waiting:
      return <Navigate to={`/raffles/${raffleAddress}/waiting`}></Navigate>;
    case RaffleState.Conditions:
      return <Navigate to={`/raffles/${raffleAddress}/conditions`}></Navigate>;
    case RaffleState.Timer:
      return <Navigate to={`/raffles/${raffleAddress}/timer`}></Navigate>;
    case RaffleState.Participation:
      return <Navigate to={`/raffles/${raffleAddress}/participation`}></Navigate>;
    case RaffleState.Result:
      return <Navigate to={`/raffles/${raffleAddress}/result`}></Navigate>;
  }

  return <></>;
};
