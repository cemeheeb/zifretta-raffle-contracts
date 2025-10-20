import {Block} from "konsta/react";
import {useParams} from "react-router-dom";
import {useContext, useMemo} from "react";
import {BlockchainContext} from "@contexts";
import {RaffleProgressStep} from "@workers";
import {Timeline} from "@components";
import {useNavigationBar} from "@hooks";

import {NAVIGATION_ITEMS} from "../constants";

interface Props {
  step: RaffleProgressStep;
}

const targetPaths = ['/raffles'];

export default ({step}: Props) => {
  const { raffleAddress } = useParams();
  const blockchainData = useContext(BlockchainContext);

  useNavigationBar(targetPaths, NAVIGATION_ITEMS);

  const steps = useMemo(() => {
  }, [step]);

  return <>
    <Block>
      <Timeline/>
    </Block>
  </>;
};

