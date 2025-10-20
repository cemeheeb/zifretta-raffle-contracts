import {FC, useState} from "react";
import {Outlet} from "react-router-dom";
import {Page, Block} from "konsta/react";
import {TonConnectButton} from "@tonconnect/ui-react";

import {NavigationContextProvider} from "@contexts";
import {NavigationBar} from "@components";

export const NavigationLayout: FC = () => {
  const [tabs, setTabs] = useState([]);

  return <NavigationContextProvider value={setTabs}>
    <Page>
      <Block className="flex justify-end me-5">
        <TonConnectButton className="flex"/>
      </Block>
      <Outlet/>
      <NavigationBar tabs={tabs}/>
    </Page>
  </NavigationContextProvider>;
};
