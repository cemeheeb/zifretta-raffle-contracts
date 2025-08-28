import {Blockchain, SandboxContract, TreasuryContract} from '@ton/sandbox';
import {Cell, toNano} from '@ton/core';
import '@ton/test-utils';
import {compile, sleep} from '@ton/blueprint';

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
              conditions: {
                blackTicketPurchases: 2,
                whiteTicketMints: 2
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
    const participantQuantity = await raffle.getParticipantQuantity();
    expect(participantQuantity).toBe(0);
  });

  it('should deploy RaffleCandidate contracts after sendRegisterCandidate', async () => {

    for (const user of users) {

      const registerCandidateResult = await raffle.sendRegisterCandidate(user.getSender(), {value: toNano("0.1")});
      const raffleCandidateAddress = await raffle.getRaffleCandidateAddress(user.address);
      const raffleCandidate: SandboxContract<RaffleCandidate> = blockchain.openContract(
          RaffleCandidate.createFromAddress(raffleCandidateAddress)
      );

      const userAddress = await raffleCandidate.getUserAddress();
      const participantIndex = await raffleCandidate.getParticipantIndex();

      expect(participantIndex).toBeNull();
      expect(userAddress.toRawString()).toBe(user.address.toRawString());

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
      const registerCandidateResult = await raffle.sendRegisterCandidate(user.getSender(), {value: toNano("0.1")});
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
        value: toNano("0.1"),
        userAddress: user.address,
        conditions: {
          blackTicketPurchases: 1,
          whiteTicketMints: 1
        }
      });

      expect(setConditionsAResult.transactions).toHaveTransaction({
        from: deployer.address,
        to: raffle.address,
        op: OperationCodes.OP_RAFFLE_SET_CONDITIONS
      });

      expect(setConditionsAResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        op: OperationCodes.OP_RAFFLE_CANDIDATE_SET_CONDITIONS
      });

      const setConditionsBResult = await raffle.sendConditions(deployer.getSender(), {
        value: toNano("0.1"),
        userAddress: user.address,
        conditions: {
          blackTicketPurchases: 2,
          whiteTicketMints: 2
        }
      });

      expect(setConditionsBResult.transactions).toHaveTransaction({
        from: raffleCandidate.address,
        to: raffle.address,
        op: OperationCodes.OP_RAFFLE_APPROVE
      });

      expect(setConditionsBResult.transactions).toHaveTransaction({
        from: raffle.address,
        to: raffleCandidate.address,
        op: OperationCodes.OP_RAFFLE_CANDIDATE_SET_PARTICIPANT_INDEX
      });

      // Должен увеличится счетчик аппрувнутых участников
      expect(await raffle.getParticipantQuantity()).toBe(++userIndex);

      const participantIndex = await raffleCandidate.getParticipantIndex();
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
  });
});
