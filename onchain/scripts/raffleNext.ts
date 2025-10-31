import {toNano} from '@ton/core';
import {NetworkProvider} from '@ton/blueprint';

import {Raffle} from '../wrappers/Raffle';
import {promptAmount} from "../wrappers/ui-utils";

export async function run(provider: NetworkProvider, args: string[]) {
  const ui = provider.ui();

  const address = await ui.inputAddress('Raffle address');

  if (!(await provider.isContractDeployed(address))) {
    ui.write(`Error: Raffle contract at address ${address} is not deployed!`);
    return;
  }

  const amount = await promptAmount('Winning amount', ui);

  const raffle = provider.open(Raffle.createFromAddress(address));

  await raffle.sendRaffleNext(provider.sender(), {
    value: amount + toNano("0.05"),
    forwardAmount: amount,
    message: "Congrats! You are winner!"
  });

  ui.clearActionPrompt();
  ui.write("New user condition was pushed into blockchain");
}
