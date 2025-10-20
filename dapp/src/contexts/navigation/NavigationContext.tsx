import {createContext} from "react";

import {NavigationTab} from "@/contexts/navigation/types.ts";

export const NavigationContext = createContext<(context: NavigationTab[]) => void>(() => {});

export const NavigationContextProvider = NavigationContext.Provider;
