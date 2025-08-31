import {Blockchain, SandboxContract, TreasuryContract} from '@ton/sandbox';
import {beginCell, Cell, toNano} from '@ton/core';
import '@ton/test-utils';
import {compile} from '@ton/blueprint';

import {OperationCodes, Raffle} from '../wrappers/Raffle';
import {RaffleCandidate} from "../wrappers/RaffleCandidate";
import {RaffleParticipant} from "../wrappers/RaffleParticipant";

const USER_QUANTITY = 5;

describe('Raffle', () => {
  let code: Cell;

  beforeAll(async () => {
    code = await compile('Raffle');
  });

  let blockchain: Blockchain;
  let deployer: SandboxContract<TreasuryContract>;
  let users: Array<SandboxContract<TreasuryContract>>;
  let raffle: SandboxContract<Raffle>;

  beforeEach(async () => {
    blockchain = await Blockchain.create();

    deployer = await blockchain.treasury('deployer');
    users = [];
    for (let i = 0; i < USER_QUANTITY; i++) {
      users.push(await blockchain.treasury('user_' + i));
    }

    raffle = blockchain.openContract(
        await Raffle.createFromConfig(
            {
              ownerAddress: deployer.address,
              deadline: BigInt(jest.now() + 100000),
              maxRewards: 3n,
              conditions: {
                blackTicketPurchased: BigInt(2),
                whiteTicketMinted: BigInt(2)
              }
            },
            code
        )
    );

    const deployResult = await raffle.sendDeploy(deployer.getSender(), toNano('0.05'));

    expect(deployResult.transactions).toHaveTransaction({
      from: deployer.address,
      to: raffle.address,
      deploy: true,
      success: true,
    });
  });

  it('should deploy', async () => {
    // the check is done inside beforeEach
    // blockchain and main are ready to use
  });

  it('participant quantity must be zero', async () => {
    const staticData = await raffle.getStaticData();
    expect(staticData.participantsQuantity).toBe(0n);
  });

  it('should deploy RaffleCandidate contracts after sendRegisterCandidate', async () => {

    for (const user of users) {

      const registerCandidateResult = await raffle.sendRegisterCandidate(user.getSender(), {
        value: toNano("0.2"),
        telegramID: 123n
      });
      const raffleCandidateAddress = await raffle.getRaffleCandidateAddress(user.address);
      const raffleCandidate: SandboxContract<RaffleCandidate> = blockchain.openContract(
          RaffleCandidate.createFromAddress(raffleCandidateAddress)
      );

      const participantIndex = await raffleCandidate.getParticipantIndex();

      expect(participantIndex).toBeNull();

      expect(registerCandidateResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        deploy: true,
        success: true,
      });
    }
  });

  it('should deploy RaffleParticipant contract on conditions reached', async () => {

    let userIndex = 0;
    for (const user of users) {
      const registerCandidateResult = await raffle.sendRegisterCandidate(user.getSender(), {
        value: toNano("0.2"),
        telegramID: 123n
      });
      const raffleCandidateAddress = await raffle.getRaffleCandidateAddress(user.address);
      const raffleCandidate: SandboxContract<RaffleCandidate> = blockchain.openContract(
          RaffleCandidate.createFromAddress(raffleCandidateAddress)
      );

      expect(registerCandidateResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        deploy: true,
        success: true,
      });

      // Отправка части
      const setConditionsAResult = await raffle.sendConditions(deployer.getSender(), {
        value: toNano("0.2"),
        userAddress: user.address,
        conditions: {
          blackTicketPurchased: 1n,
          whiteTicketMinted: 1n
        }
      });

      expect(setConditionsAResult.transactions).toHaveTransaction({
        from: deployer.address,
        to: raffle.address,
        op: OperationCodes.OP_RAFFLE_SET_CONDITIONS,
        success: true
      });

      expect(setConditionsAResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        op: OperationCodes.OP_RAFFLE_CANDIDATE_SET_CONDITIONS,
        success: true
      });

      expect(setConditionsAResult.transactions).toHaveTransaction({
        from: raffleCandidate.address,
        to: deployer.address,
        success: true
      });

      const setConditionsBResult = await raffle.sendConditions(deployer.getSender(), {
        value: toNano("0.2"),
        userAddress: user.address,
        conditions: {
          blackTicketPurchased: 2n,
          whiteTicketMinted: 2n
        }
      });

      expect(setConditionsBResult.transactions).toHaveTransaction({
        from: raffleCandidate.address,
        to: raffle.address,
        op: OperationCodes.OP_RAFFLE_APPROVE,
        success: true
      });

      expect(setConditionsBResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        op: OperationCodes.OP_RAFFLE_CANDIDATE_SET_PARTICIPANT_INDEX,
        success: true
      });

      // Должен увеличится счетчик аппрувнутых участников
      const staticData = await raffle.getStaticData();
      expect(staticData.participantsQuantity).toBe(BigInt(++userIndex));

      const participantIndex = await raffleCandidate.getParticipantIndex();
      console.warn({participantIndex});
      if (participantIndex != null) {
        const raffleParticipantAddress = await raffle.getRaffleParticipantAddress(participantIndex!);
        let raffleParticipant: SandboxContract<RaffleParticipant> = blockchain.openContract(
            RaffleParticipant.createFromAddress(raffleParticipantAddress)
        );

        // Должен быть задеплоен контракт аппрувнутого участника,
        expect(setConditionsBResult.transactions).toHaveTransaction({
          from: raffle.address,
          to: raffleParticipant!.address,
          deploy: true,
          success: true,
        });
      }

      expect(participantIndex).toBeDefined();
    }

    console.warn(await raffle.getRaffleParticipantAddress(2));

    const staticData = await raffle.getStaticData();
    expect(staticData.participantsQuantity).toBe(5n);

    const setRaffleNextAResult = await raffle.sendRaffleNext(deployer.getSender(), {
      value: toNano("0.2"),
      message: "Congratulations! You are the winner!"
    });

    expect(setRaffleNextAResult.transactions).toHaveTransaction({
      success: true,
      body: beginCell().storeMaybeStringRefTail("Congratulations! You are the winner!").endCell().beginParse().loadRef()
    });
  });
});
