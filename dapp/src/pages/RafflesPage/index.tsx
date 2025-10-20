import {Block} from "konsta/react";
import {useMemo} from "react";
import {useNavigationBar} from "@hooks";
import {NAVIGATION_ITEMS} from "@/pages/constants.ts";

export default () => {

  const targetPaths = useMemo(() => {
    return ["/participants"];
  }, []);

  useNavigationBar(targetPaths, NAVIGATION_ITEMS);

  return <>
    <Block>
      PARTICIPANTS PAGE
    </Block>
  </>;
};

