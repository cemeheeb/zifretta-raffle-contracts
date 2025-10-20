import {useMemo, useState, useEffect} from "react";
import {
  createBrowserRouter,
  createRoutesFromElements,
  Navigate,
  Route,
  RouterProvider
} from "react-router-dom";
import {MainButton} from "@vkruglikov/react-telegram-web-app";
import {useTonAddress} from "@tonconnect/ui-react";
import {App, KonstaProvider} from "konsta/react";

import {NavigationLayout} from "@layouts";
import {
  RafflesPage,
  ProfilePage,
  RaffleDetailQualificationPage,
  RaffleDetailTimelinePage,
  RaffleDetailRafflePage,
  RaffleDetailsStateRedirectPage
} from "@pages";
import {useBlockchain} from "@hooks";

import {BlockchainData, RaffleProgressStep} from "@workers";
import {BlockchainContextProvider} from "@contexts";

import './index.css';
import {} from "@pages";

const ORACLE_ADDRESS = "UQBIxYxdmJHYs7GT3nJpIKy-oor4n3WBuf93weCIGjgN4yxr";

export default () => {

  const userAddress = useTonAddress(true);
  const {getBlockchainData} = useBlockchain();

  const [blockchainData, setBlockchainData] = useState<BlockchainData>({
    raffles: [],
  });

  useEffect(() => {
    if (!userAddress) {
      return;
    }

    getBlockchainData(ORACLE_ADDRESS, userAddress).then((data) => {
      console.log('setBlockchainData', data);
      setBlockchainData(data);
    });

    return () => {
    }
  }, [userAddress]);

  const router = useMemo(() => {
    return createBrowserRouter(
      createRoutesFromElements(
        <Route element={<NavigationLayout/>}>
          <Route index element={<Navigate to="raffles" replace/>}/>
          <Route path="raffles">
            <Route index element={<Navigate to="launched" replace/>}/>
            <Route path="launched" element={<RafflesPage segment="launched"/>}/>
            <Route path="completed" element={<RafflesPage segment="completed"/>}/>
            <Route path=":raffleAddress">
              <Route index element={<RaffleDetailsStateRedirectPage />}/>
              <Route path="qualification" element={<RaffleDetailQualificationPage registration/>}/>
              <Route path="waiting" element={<RaffleDetailTimelinePage step={RaffleProgressStep.Waiting}/>}/>
              <Route path="conditions" element={<RaffleDetailQualificationPage/>}/>
              <Route path="timer" element={<RaffleDetailTimelinePage step={RaffleProgressStep.Timer}/>}/>
              <Route path="participation" element={<RaffleDetailTimelinePage step={RaffleProgressStep.Participation}/>}/>
              <Route path="winners" element={<RaffleDetailRafflePage/>}/>
            </Route>
          </Route>
          <Route path="profile" element={<ProfilePage/>}/>
        </Route>
      ));
  }, []);

  return (<>
      <KonstaProvider>
        <App theme="material" dark={false}>
          <BlockchainContextProvider value={blockchainData}>
            <RouterProvider router={router}/>
          </BlockchainContextProvider>
        </App>
      </KonstaProvider>
      <MainButton/>
    </>
  );
};
