import dayjs, {Dayjs} from "dayjs";

const formatInternal = (value: Dayjs | string, format: string) => {
  if (!value) {
    return '';
  }

  if (typeof value === 'string') {
    value = dayjs(value);
  }

  return value.format(format);
}
export const formatDayMonth = (value: Dayjs) => formatInternal(value, "DD MMMM");
export const formatDate = (value: Dayjs) => formatInternal(value, "DD.MM.YYYY");
export const formatDateTime = (value: Dayjs) => formatInternal(value, "DD.MM.YYYY HH:mm:ss");
export const formatTime = (value: Dayjs) => formatInternal(value, "HH:mm:ss");
