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

  for (let i = 0; i < 25; i++) {
    ui.write("sending raffle next message");

    await raffle.sendRaffleNext(provider.sender(), {
      value: toNano("0.01") + toNano("0.05"),
      forwardAmount: toNano("0.01"),
      message: "Congrats! You are winner!"
    });

    await new Promise(resolve => setTimeout(resolve, 3000));
  }

  ui.clearActionPrompt();
  ui.write("New user condition was pushed into blockchain");
}
