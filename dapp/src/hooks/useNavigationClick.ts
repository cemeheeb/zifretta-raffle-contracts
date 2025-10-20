import {useCallback} from "react";
import {useNavigate} from "react-router-dom";

export const useNavigationClick = () => {
  const navigate = useNavigate();

  return useCallback((link: string) => {
    navigate(link, {replace: true});
  }, [navigate]);
}
