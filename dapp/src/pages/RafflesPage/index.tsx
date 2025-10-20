import {Block, Segmented, SegmentedButton} from "konsta/react";
import {useMemo} from "react";

import {useNavigationBar, useNavigationClick} from "@hooks";

import {NAVIGATION_ITEMS} from "../constants.ts";
import RafflesPageLaunchedSegment from "./RafflesPageLaunchedSegment";
import RafflesPageCompletedSegment from "./RafflesPageCompletedSegment";
import {useLocation} from "react-router-dom";

type Props = {
  segment: 'launched' | 'completed';
}

const targetPaths = ["/raffles/launched", "/raffles/completed"];

export default ({segment}: Props) => {
  const location = useLocation();

  const navigationClick = useNavigationClick();
  useNavigationBar(targetPaths, NAVIGATION_ITEMS);

  const segmentElement = useMemo(() => {
    switch (segment) {
      case 'launched': {
        return <RafflesPageLaunchedSegment />;
      }
      case 'completed': {
        return <RafflesPageCompletedSegment />;
      }
    }
  }, [
    segment,
  ]);

  return <>
    <Block>
      <Segmented strong rounded>
        <SegmentedButton strong rounded active={location.pathname === "/raffles/launched"}
                         onClick={() => navigationClick("/raffles/launched")}>
          Актуальные
        </SegmentedButton>
        <SegmentedButton strong rounded active={location.pathname === "/raffles/completed"}
                         onClick={() => navigationClick("/raffles/completed")}>
          Завершенные
        </SegmentedButton>
      </Segmented>
    </Block>
    <Block>
      {segmentElement}
    </Block>
  </>;
};

