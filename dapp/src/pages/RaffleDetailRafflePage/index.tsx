import {Block, Card, List, ListItem} from "konsta/react";
import {useParams} from "react-router-dom";

interface Props {
  step: 'conditions' | 'waiting' | '';
}

export default () => {
  const { raffleAddress } = useParams();

  return <>
    <Block>
      RAFFLE STATUS PAGE
      <Card>
        <List title="Для участия выполните следующие условия:">
          {/*<ListItem title={`Черных тикетов приобретено через getgems: ${userConditions.blackTicketPurchased} из ${raffleConditions.blackTicketPurchased}`} />*/}
          {/*<ListItem title={`Белых тикетов сминчено: ${userConditions.whiteTicketMinted} из ${raffleConditions.whiteTicketMinted}`} />*/}
        </List>
      </Card>
    </Block>
  </>;
};

