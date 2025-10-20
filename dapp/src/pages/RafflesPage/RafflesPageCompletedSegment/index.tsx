import {Block, Button} from "konsta/react";
import {useMemo, useContext} from "react";
import {useNavigationBar, useNavigationClick} from "@hooks";
import {NAVIGATION_ITEMS} from "@/pages/constants.ts";
import {BlockchainContext} from "@contexts";
import {selectLaunchedRafflesData, selectCompletedRafflesData} from "@workers";
import {Navigate} from "@assets/icons";

const targetPaths = ["/raffles/completed"];

export default () => {

  const navigationClick = useNavigationClick();
  useNavigationBar(targetPaths, NAVIGATION_ITEMS);

  const blockchainData = useContext(BlockchainContext);
  const raffles = useMemo(() => selectCompletedRafflesData(blockchainData), [blockchainData]);
  console.log("blockchainData", blockchainData);

  return <>
    {
      raffles.map(raffleData => {
        return <Button className="flex flex-row justify-between items-center" large onClick={() => navigationClick("/raffles/" + raffleData.address)}>
          {raffleData.address}
          <Navigate width={16} height={16} />
        </Button>
      })
    }
  </>;
};

