import {toNano} from '@ton/core';
import {NetworkProvider} from '@ton/blueprint';

import {Raffle} from '../wrappers/Raffle';

export async function run(provider: NetworkProvider, args: string[]) {
  const ui = provider.ui();

  const address = await ui.inputAddress('Raffle address');

  if (!(await provider.isContractDeployed(address))) {
    ui.write(`Error: Raffle contract at address ${address} is not deployed!`);
    return;
  }

  const raffle = provider.open(Raffle.createFromAddress(address));

  await raffle.sendRaffleNext(provider.sender(), {
    value: toNano("0.3"),
    forwardAmount: toNano("0.15"),
    message: "Congrats! You are winner!"
  });

  ui.clearActionPrompt();
  ui.write("New user condition was pushed into blockchain");
}
