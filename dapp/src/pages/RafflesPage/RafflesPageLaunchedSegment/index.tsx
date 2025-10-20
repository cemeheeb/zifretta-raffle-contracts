import {Button} from "konsta/react";
import {useContext, useMemo} from "react";
import {Navigate} from '@assets/icons'

import {useNavigationClick, useNavigationBar} from "@hooks";
import {BlockchainContext} from "@contexts";
import {selectLaunchedRafflesData} from "@workers";
import {NAVIGATION_ITEMS} from "@/pages/constants.ts";

export default () => {

  const targetPaths = useMemo(() => {
    return ["/raffles/launched"];
  }, []);

  useNavigationBar(targetPaths, NAVIGATION_ITEMS);
  const navigationClick = useNavigationClick();

  const blockchainData = useContext(BlockchainContext);
  const raffles = useMemo(() => selectLaunchedRafflesData(blockchainData), [blockchainData]);
  console.log("blockchainData", raffles);

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

