const regex = /^([a-zA-Z0-9-_]{20})[a-zA-Z0-9-_]+([a-zA-Z0-9-_]{4})$/;

export const truncateTonAddress = (address: string) => {
  if (!address) {
    return '';
  }

  const match = address.match(regex);
  if (!match) return address;
  return `${match[1]}…${match[2]}`;
};
