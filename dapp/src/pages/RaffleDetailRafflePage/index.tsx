import {Block, Card, List, ListItem} from "konsta/react";
import {useParams} from "react-router-dom";

export default () => {
  const { raffleAddress } = useParams();

  return <>
    <Block>
      RAFFLE
    </Block>
  </>;
};

