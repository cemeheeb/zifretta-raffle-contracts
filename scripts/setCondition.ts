import {Address, toNano} from '@ton/core';
import {NetworkProvider} from '@ton/blueprint';

import {Raffle} from '../wrappers/Raffle';
import {promptBigint} from "../wrappers/ui-utils";

export async function run(provider: NetworkProvider, args: string[]) {
  const ui = provider.ui();

  const address = await ui.inputAddress('Raffle address');
  const blackTicketPurchased = await promptBigint('Amount of black ticket purchased', ui);
  const whiteTicketMinted = await promptBigint('Amount of white ticket minted', ui);

  if (!(await provider.isContractDeployed(address))) {
    ui.write(`Error: Contract at address ${address} is not deployed!`);
    return;
  }

  const raffle = provider.open(Raffle.createFromAddress(address));

  await raffle.sendConditions(provider.sender(), {
    value: toNano("0.01"),
    userAddress: Address.parse(""),
    conditions: {
      blackTicketPurchased,
      whiteTicketMinted
    }
  });

  ui.clearActionPrompt();
  ui.write("New user condition was pushed into blockchain");
}
