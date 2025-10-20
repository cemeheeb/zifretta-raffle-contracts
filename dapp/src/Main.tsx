import React from "react";
import ReactDOM from "react-dom/client";
import {WebAppProvider} from "@vkruglikov/react-telegram-web-app";

import Application from "@/Application.tsx";
import {TonConnectUIProvider} from "@tonconnect/ui-react";

ReactDOM.createRoot(
  document.getElementById("root")!).render(
  <TonConnectUIProvider manifestUrl="https://nft.zifretta.com/manifest.json">
    <WebAppProvider options={{
      smoothButtonsTransition: true,
    }}>
      <Application/>
    </WebAppProvider>
  </TonConnectUIProvider>
);
