import {Block, Button} from "konsta/react";
import {useParams} from "react-router-dom";
import {useContext, useMemo, useState} from "react";
import {BlockchainContext} from "@contexts";
import {selectRaffle} from "@workers";
import {useTonAddress, useTonConnectModal} from "@tonconnect/ui-react";

interface Props {
  registration?: boolean;
}

export default ({registration = false}: Props) => {
  const {raffleAddress} = useParams();

  const {open} = useTonConnectModal();
  const address = useTonAddress();
  const blockchainData = useContext(BlockchainContext);

  const [isRegisterClicked, setIsRegisterClicked] = useState(false);

  const raffle = selectRaffle(blockchainData, raffleAddress);

  const button = useMemo(() => {

    if (raffle?.raffleCandidateData || isRegisterClicked) {
      return null;
    }

    const onConnectClick = () => {
      open();
    }

    const onRegisterClick = () => {
      setIsRegisterClicked(true);
    }

    return address === ""
      ? <Button small onClick={onConnectClick}>Подключить кошелек</Button>
      : <Button small onClick={onRegisterClick}>Зарегистрироваться</Button>;
  }, [address, open, raffle?.raffleCandidateData, isRegisterClicked]);

  return <>
    <Block>
      <div className="flex flex-row gap-2">
        <Button small>Условия</Button>
        <Button small>Поддержка</Button>
      </div>
      <div className="flex flex-col gap-2 mt-2">
        {button}
        <Button small
                disabled={registration}>{`1. Приобрести NFT Black Ticket ${raffle?.raffleCandidateData?.conditions?.blackTicketPurchased ?? 0}/${raffle?.raffleData.conditions.blackTicketPurchased}`}</Button>
        <Button small
                disabled={registration}>{`2. Минт NFT White Ticket ${raffle?.raffleCandidateData?.conditions?.whiteTicketMinted ?? 0}/${raffle?.raffleData.conditions.whiteTicketMinted}`}</Button>
      </div>
    </Block>
  </>;
};

