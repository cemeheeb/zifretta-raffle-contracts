import {Address, toNano} from '@ton/core';
import {Raffle} from '../wrappers/Raffle';
import {compile, NetworkProvider} from '@ton/blueprint';
import {promptBigint} from "../wrappers/ui-utils";

export async function run(provider: NetworkProvider) {
  const ui = provider.ui();

  let minParticipantsQuantity: bigint = await promptBigint('Minimum candidates quantity:', ui);
  let duration: bigint = await promptBigint('Participant registration await duration in seconds:', ui);

  const raffle = provider.open(
      await Raffle.createFromConfig(
          {
            ownerAddress: provider.sender().address!,
            minParticipantsQuantity,
            duration, // 7 days in seconds = 604800n
            conditions: {
              blackTicketPurchased: 2n,
              whiteTicketMinted: 2n
            }
          },
          await compile('Raffle')
      )
  );

  await raffle.sendDeploy(provider.sender(), toNano('0.03'));
  await provider.waitForDeploy(raffle.address);
}
