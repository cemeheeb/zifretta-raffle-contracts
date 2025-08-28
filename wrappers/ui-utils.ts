import {UIProvider} from "@ton/blueprint";
import {toNano} from "@ton/core";

export const promptBigint = async (prompt: string, provider: UIProvider) => {
  let resAmount: number;
  do {
    let inputAmount = await provider.input(prompt);
    resAmount = parseInt(inputAmount);
    if (isNaN(resAmount)) {
      provider.write("Cannot convert '" + inputAmount + "' string to number");
    } else {
      return BigInt(Math.floor(resAmount)).valueOf();
    }
  } while (true);
}

export const promptAmount = async (prompt: string, provider: UIProvider) => {
  let resAmount: bigint;
  do {
    let inputAmount = await provider.input(prompt);
    try {
      resAmount = toNano(inputAmount);
      return resAmount.valueOf();
    } catch (error) {
      provider.write("Cannot convert '" + inputAmount + "' string to coins");
    }
  } while (true);
}