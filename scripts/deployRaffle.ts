import {Address, toNano} from '@ton/core';
import {Raffle} from '../wrappers/Raffle';
import {compile, NetworkProvider} from '@ton/blueprint';

export async function run(provider: NetworkProvider) {
  const ui = provider.ui();

  let deadline: bigint | null = null;
  do {
    const inputText = await ui.input('Deadline in unix timestamp in seconds:');
    const parsed = parseInt(inputText);
    if (isNaN(parsed)) {
      continue;
    }
    deadline = BigInt(parsed);
  } while (deadline == null)

  const raffle = provider.open(
      await Raffle.createFromConfig(
          {
            ownerAddress: provider.sender().address!,
            deadline,
          },
          await compile('Raffle')
      )
  );

  await raffle.sendDeploy(provider.sender(), toNano('0.02'));
  await provider.waitForDeploy(raffle.address);

  console.log('deadline', await raffle.getDeadline());
}
