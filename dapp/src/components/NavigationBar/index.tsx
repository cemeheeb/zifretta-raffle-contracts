import {useCallback} from "react";
import {useLocation, useNavigate} from 'react-router-dom';
import {NavigationTab} from "@contexts";
import {Tabbar, TabbarLink} from 'konsta/react'

interface Props {
  tabs: NavigationTab[];
}

export const NavigationBar = ({tabs}: Props) => {
  const location = useLocation();
  const navigate = useNavigate();

  const onNavigationClick = useCallback((link: string) => {
    navigate(link, {replace: true});
  }, [navigate]);

  return (
    <Tabbar className="bottom-0 fixed" icons>
      {tabs.map(({text, path}) => {
        return <TabbarLink key={path}
                           onClick={() => onNavigationClick(path)}
                           active={location.pathname.indexOf(path) >= 0}
        >
          {text}
        </TabbarLink>
      })}
    </Tabbar>
  )
};
