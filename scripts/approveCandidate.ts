import {Address, toNano} from '@ton/core';
import {Raffle} from '../wrappers/Raffle';
import {NetworkProvider, sleep} from '@ton/blueprint';

export async function run(provider: NetworkProvider, args: string[]) {
  const ui = provider.ui();

  const address = Address.parse(args.length > 0 ? args[0] : await ui.input('Raffle address'));

  if (!(await provider.isContractDeployed(address))) {
    ui.write(`Error: Contract at address ${address} is not deployed!`);
    return;
  }

  const raffle = provider.open(Raffle.createFromAddress(address));
  const participantQuantityBefore = await raffle.getParticipantQuantity();

  await raffle.sendApproveCandidate(provider.sender(), {
    value: toNano("0.01"),
    userAddress: Address.parse(""),
  });

  ui.clearActionPrompt();
  ui.write("Congrats, you're successfully registered!");
}
