import {useContext, useEffect} from "react";
import {useLocation} from "react-router-dom";
import {NavigationContext, NavigationTab} from "@contexts";

export function useNavigationBar(targetPaths: string[], tabs: NavigationTab[]) {
  const location = useLocation();
  const setTabs = useContext(NavigationContext);

  if (!setTabs) {
    throw new Error("Could not use NavigationContext outside of NavigationContextProvider");
  }

  useEffect(() => {
    if (targetPaths.indexOf(location.pathname) < 0) {
      return;
    }

    setTabs(tabs);
  }, [location.pathname]);
}
