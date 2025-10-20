import {createContext} from "react";

import {NavigationTab} from "./types";

export const NavigationContext = createContext<(context: NavigationTab[]) => void>(() => {});

export const NavigationContextProvider = NavigationContext.Provider;
