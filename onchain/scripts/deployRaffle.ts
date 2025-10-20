import {Address, toNano} from '@ton/core';
import {Raffle} from '../wrappers/Raffle';
import {compile, NetworkProvider} from '@ton/blueprint';
import {promptBigint} from "../wrappers/ui-utils";

export async function run(provider: NetworkProvider) {
  const ui = provider.ui();

  let minCandidateQuantity: bigint = await promptBigint('Minimum candidates quantity:', ui);

  const raffle = provider.open(
      await Raffle.createFromConfig(
          {
            ownerAddress: provider.sender().address!,
            minParticipantsQuantity: minCandidateQuantity,
            duration: 604800n, // 7 days in seconds
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
