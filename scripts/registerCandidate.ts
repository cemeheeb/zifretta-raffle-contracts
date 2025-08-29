import { toNano } from '@ton/core';
import { Raffle } from '../wrappers/Raffle';
import { NetworkProvider } from '@ton/blueprint';
import {promptBigint} from "../wrappers/ui-utils";

export async function run(provider: NetworkProvider) {
  const ui = provider.ui();

  const address = await ui.inputAddress('Raffle address');
  if (!(await provider.isContractDeployed(address))) {
    ui.write(`Error: Contract at address ${address} is not deployed!`);
    return;
  }

  const telegramID = await promptBigint('Telegram id:', ui);

  const raffle = provider.open(Raffle.createFromAddress(address));
  await raffle.sendRegisterCandidate(provider.sender(), {
    value: toNano("0.003"),
    telegramID
  });

  ui.write('Waiting for candidate registration...');

  ui.clearActionPrompt();
  ui.write('Candidate successfully registered!');
}
