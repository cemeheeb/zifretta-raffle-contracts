import {Block, Button, BlockHeader} from "konsta/react";
import {useMemo} from "react";
import {useNavigationBar} from "@hooks";
import {NAVIGATION_ITEMS} from "@/pages/constants.ts";

export default () => {

  const targetPaths = useMemo(() => {
    return ["/profile"];
  }, []);

  useNavigationBar(targetPaths, NAVIGATION_ITEMS);

  return <>
    <Block strong inset>
      Semeneev Eldar
    </Block>
    <Block>
      <Button large>Отключить кошелек</Button>
    </Block>
    <div className="flex flex-row justify-around">
      <Block margin="5px" className="flex-1" strong inset>
        <p className="text-gray-400">Приняли участие:</p>
        <span className="text-4xl">120</span>
      </Block>
      <Block margin="5px" className="flex-1 m-0" strong inset>
        <p className="text-gray-400">История призов:</p>
        <span className="text-4xl">120</span>
      </Block>
    </div>
  </>;
};

