import {MouseEventHandler} from "react";

export const stopPropagation: MouseEventHandler<HTMLDivElement>  = (event) => {
  event.stopPropagation();
}
