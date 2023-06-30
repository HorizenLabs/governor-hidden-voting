const { constants, expectEvent, expectRevert } = require('@openzeppelin/test-helpers');
const { expect } = require('chai');
const { ProposalState } = require('../../lib/openzeppelin-contracts/test/helpers/enums');
const { GovernorEncryptedHelper, VoteType } = require('../helpers/governance.encrypted');
const { clockFromReceipt } = require('../../lib/openzeppelin-contracts/test/helpers/time');
const { loadEVotingBackend } = require('../../../backend/wasm/assets/wasm_exec_node');

const Governor = artifacts.require('$GovernorEncryptedMock');
const CallReceiver = artifacts.require('CallReceiverMock');

const TOKENS = [
  { Token: artifacts.require('$ERC20VotesMock'), mode: 'blocknumber' },
];

contract('GovernorEncrytped', function (accounts) {
  const [owner, proposer, voter1, voter2, voter3, voter4] = accounts;

  const name = 'OZ-Governor';
  const tokenName = 'MockToken';
  const tokenSymbol = 'MTKN';
  const tokenSupply = web3.utils.toWei('100', 'wei');
  const votingDelay = web3.utils.toBN(4);
  const votingPeriod = web3.utils.toBN(16);
  const tallyingPeriod = web3.utils.toBN(4);
  const value = web3.utils.toWei('1', 'wei');

  for (const { mode, Token } of TOKENS) {
    describe(`using ${Token._json.contractName}`, function () {
      beforeEach(async function () {
        await loadEVotingBackend();
        this.chainId = await web3.eth.getChainId();
        this.token = await Token.new(tokenName, tokenSymbol, tokenName);
        this.mock = await Governor.new(
          name, // name
          this.token.address, // tokenAddress
          10, // quorumNumeratorValue
          votingDelay, // initialVotingDelay
          votingPeriod, // initialVotingPeriod
          0, // initialProposalThreshold
          tallyingPeriod, // initialTallyingPeriod
        );
        this.receiver = await CallReceiver.new();

        this.helper = new GovernorEncryptedHelper(this.mock, mode);

        await web3.eth.sendTransaction({ from: owner, to: this.mock.address, value });

        await this.token.$_mint(owner, tokenSupply);
        await this.helper.delegate({ token: this.token, to: voter1, value: web3.utils.toWei('10', 'wei') }, { from: owner });
        await this.helper.delegate({ token: this.token, to: voter2, value: web3.utils.toWei('7', 'wei') }, { from: owner });
        await this.helper.delegate({ token: this.token, to: voter3, value: web3.utils.toWei('5', 'wei') }, { from: owner });
        await this.helper.delegate({ token: this.token, to: voter4, value: web3.utils.toWei('2', 'wei') }, { from: owner });

        this.proposal = this.helper.setProposal(
          [
            {
              target: this.receiver.address,
              data: this.receiver.contract.methods.mockFunction().encodeABI(),
              value,
            },
          ],
          '<proposal description>',
        );
      });

      it('deployment check', async function () {
        expect(await this.mock.name()).to.be.equal(name);
        expect(await this.mock.token()).to.be.equal(this.token.address);
        expect(await this.mock.votingDelay()).to.be.bignumber.equal(votingDelay);
        expect(await this.mock.votingPeriod()).to.be.bignumber.equal(votingPeriod);
        expect(await this.mock.quorum(0)).to.be.bignumber.equal('0');
        expect(await this.mock.COUNTING_MODE()).to.be.equal('support=ignore&quorum=bravo&params=bn256helios');
      });

      it('nominal workflow', async function () {
        // Before
        expect(await this.mock.proposalProposer(this.proposal.id)).to.be.equal(constants.ZERO_ADDRESS);
        expect(await this.mock.hasVoted(this.proposal.id, owner)).to.be.equal(false);
        expect(await this.mock.hasVoted(this.proposal.id, voter1)).to.be.equal(false);
        expect(await this.mock.hasVoted(this.proposal.id, voter2)).to.be.equal(false);
        expect(await web3.eth.getBalance(this.mock.address)).to.be.bignumber.equal(value);
        expect(await web3.eth.getBalance(this.receiver.address)).to.be.bignumber.equal('0');

        // Initialize pk
        const pk = await this.helper.addKeyPair();
        await this.helper.initialize({ pk });

        // Run proposal
        const txPropose = await this.helper.propose({ from: proposer });

        expectEvent(txPropose, 'ProposalCreated', {
          proposalId: this.proposal.id,
          proposer,
          targets: this.proposal.targets,
          // values: this.proposal.values,
          signatures: this.proposal.signatures,
          calldatas: this.proposal.data,
          voteStart: web3.utils.toBN(await clockFromReceipt[mode](txPropose.receipt)).add(votingDelay),
          voteEnd: web3.utils
            .toBN(await clockFromReceipt[mode](txPropose.receipt))
            .add(votingDelay)
            .add(votingPeriod),
          description: this.proposal.description,
        });

        await this.helper.waitForSnapshot();

        expectEvent(await this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
          'VoteCastWithParams',
          {
            voter: voter1,
            support: web3.utils.toBN(0),
            reason: '',
            weight: web3.utils.toWei('10', 'wei'),
          },
        );

        expectEvent(await this.helper.vote({ vote: VoteType.Against }, { from: voter2 }),
          'VoteCastWithParams',
          {
            voter: voter2,
            support: web3.utils.toBN(0),
            reason: '',
            weight: web3.utils.toWei('7', 'wei'),
          }
        );

        expectEvent(await this.helper.vote({ vote: VoteType.For }, { from: voter3 }),
          'VoteCastWithParams',
          {
            voter: voter3,
            support: web3.utils.toBN(0),
            reason: '',
            weight: web3.utils.toWei('5', 'wei'),
          }
        );

        expectEvent(await this.helper.vote({ vote: VoteType.Against }, { from: voter4 }),
          'VoteCastWithParams',
          {
            voter: voter4,
            support: web3.utils.toBN(0),
            reason: '',
            weight: web3.utils.toWei('2', 'wei'),
          }
        );

        await this.helper.waitForVotingDeadline();
        await this.helper.tally();
        await this.helper.waitForDeadline();
        const txExecute = await this.helper.execute();

        expectEvent(txExecute, 'ProposalExecuted', { proposalId: this.proposal.id });
        await expectEvent.inTransaction(txExecute.tx, this.receiver, 'MockFunctionCalled');

        // After
        expect(await this.mock.proposalProposer(this.proposal.id)).to.be.equal(proposer);
        expect(await this.mock.hasVoted(this.proposal.id, owner)).to.be.equal(false);
        expect(await this.mock.hasVoted(this.proposal.id, voter1)).to.be.equal(true);
        expect(await this.mock.hasVoted(this.proposal.id, voter2)).to.be.equal(true);
        expect(await web3.eth.getBalance(this.mock.address)).to.be.bignumber.equal('0');
        expect(await web3.eth.getBalance(this.receiver.address)).to.be.bignumber.equal(value);
      });

      it('send ethers', async function () {
        const empty = web3.utils.toChecksumAddress(web3.utils.randomHex(20));

        this.proposal = this.helper.setProposal(
          [
            {
              target: empty,
              value,
            },
          ],
          '<proposal description>',
        );

        // Before
        expect(await web3.eth.getBalance(this.mock.address)).to.be.bignumber.equal(value);
        expect(await web3.eth.getBalance(empty)).to.be.bignumber.equal('0');

        // Initialize pk
        const pk = await this.helper.addKeyPair();
        await this.helper.initialize({ pk });

        // Run proposal
        await this.helper.propose();
        await this.helper.waitForSnapshot();
        await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
        await this.helper.waitForVotingDeadline();
        await this.helper.tally();
        await this.helper.waitForDeadline();
        await this.helper.execute();

        // After
        expect(await web3.eth.getBalance(this.mock.address)).to.be.bignumber.equal('0');
        expect(await web3.eth.getBalance(empty)).to.be.bignumber.equal(value);
      });

      describe('should revert', function () {
        beforeEach(async function () {
          const pk = await this.helper.addKeyPair();
          await this.helper.initialize({ pk });
        });

        describe('on propose', function () {
          it('if proposal already exists', async function () {
            await this.helper.propose();
            await expectRevert(this.helper.propose(), 'Governor: proposal already exists');
          });
        });

        describe('on vote', function () {
          it('if proposal does not exist', async function () {
            await expectRevert(
              this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
              'Governor: unknown proposal id',
            );
          });

          it('if voting has not started', async function () {
            await this.helper.propose();
            await expectRevert(
              this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
              'Governor: vote not currently active',
            );
          });

          it('if vote value is invalid', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await expectRevert(
              this.helper.vote({ vote: 2 }),
              'GovernorCountingEncrypted: proof verification failed',
            );
          });

          it('if vote was already casted', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
            await expectRevert(
              this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
              'GovernorCountingEncrypted: vote already cast',
            );
          });

          it('if proposal is not Active', async function () {
            await this.helper.propose();
            await this.helper.waitForDeadline();
            await expectRevert(
              this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
              'Governor: vote not currently active',
            );
          });

          it('if voting is over', async function () {
            await this.helper.propose();
            await this.helper.waitForVotingDeadline();
            await expectRevert(
              this.helper.vote({ vote: VoteType.For }, { from: voter1 }),
              'GovernorCountingEncrypted: voting is over',
            );
          });
        });

        describe('on tally', function () {
          it('if proposal does not exist', async function () {
            await expectRevert(this.helper.tally({ fake: true }), 'Governor: unknown proposal id');
          });

          it('if invalid proof is sent', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter3 });
            await this.helper.waitForVotingDeadline();
            await expectRevert(
              this.helper.tally({ fake: true }),
              'GovernorCountingEncrypted: proof verification failed',
            );
          });

          it('if tallying has not started', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter3 });
            await expectRevert(
              this.helper.tally(),
              'GovernorCountingEncrypted: tallying not currently active',
            );
          });

          it('if tallying is over', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter3 });
            await this.helper.waitForDeadline();
            await expectRevert(
              this.helper.tally(),
              'GovernorCountingEncrypted: tallying not currently active',
            );
          });
        });

        describe('on execute', function () {
          it('if proposal does not exist', async function () {
            await expectRevert(this.helper.execute(), 'Governor: unknown proposal id');
          });

          it('if quorum is not reached', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter3 });
            await this.helper.waitForVotingDeadline();
            await this.helper.tally();
            await this.helper.waitForDeadline();
            await expectRevert(this.helper.execute(), 'Governor: proposal not successful');
          });

          it('if score not reached', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.Against }, { from: voter1 });
            await this.helper.waitForVotingDeadline();
            await this.helper.tally();
            await this.helper.waitForDeadline();
            await expectRevert(this.helper.execute(), 'Governor: proposal not successful');
          });

          it('if tallying has not been performed', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
            await this.helper.waitForDeadline();
            await expectRevert(this.helper.execute(), 'Governor: proposal not successful');
          });

          it('if receiver revert without reason', async function () {
            this.proposal = this.helper.setProposal(
              [
                {
                  target: this.receiver.address,
                  data: this.receiver.contract.methods.mockFunctionRevertsNoReason().encodeABI(),
                },
              ],
              '<proposal description>',
            );

            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
            await this.helper.waitForVotingDeadline();
            await this.helper.tally();
            await this.helper.waitForDeadline();
            await expectRevert(this.helper.execute(), 'Governor: call reverted without message');
          });

          it('if receiver revert with reason', async function () {
            this.proposal = this.helper.setProposal(
              [
                {
                  target: this.receiver.address,
                  data: this.receiver.contract.methods.mockFunctionRevertsReason().encodeABI(),
                },
              ],
              '<proposal description>',
            );

            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
            await this.helper.waitForVotingDeadline();
            await this.helper.tally();
            await this.helper.waitForDeadline();
            await expectRevert(this.helper.execute(), 'CallReceiverMock: reverting');
          });

          it('if proposal was already executed', async function () {
            await this.helper.propose();
            await this.helper.waitForSnapshot();
            await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
            await this.helper.waitForVotingDeadline();
            await this.helper.tally();
            await this.helper.waitForDeadline();
            await this.helper.execute();
            await expectRevert(this.helper.execute(), 'Governor: proposal not successful');
          });
        });
      });

      describe('state', function () {
        beforeEach(async function () {
          const pk = await this.helper.addKeyPair();
          await this.helper.initialize({ pk });
        });

        it('Unset', async function () {
          await expectRevert(this.mock.state(this.proposal.id), 'Governor: unknown proposal id');
        });

        it('Pending & Active', async function () {
          await this.helper.propose();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Pending);
          await this.helper.waitForSnapshot();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Pending);
          await this.helper.waitForSnapshot(+1);
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Active);
        });

        it('Defeated', async function () {
          await this.helper.propose();
          await this.helper.waitForVotingDeadline();
          await this.helper.tally();
          await this.helper.waitForDeadline();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Active);
          await this.helper.waitForDeadline(+1);
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Defeated);
        });

        it('Succeeded', async function () {
          await this.helper.propose();
          await this.helper.waitForSnapshot();
          await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
          await this.helper.waitForVotingDeadline();
          await this.helper.tally();
          await this.helper.waitForDeadline();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Active);
          await this.helper.waitForDeadline(+1);
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Succeeded);
        });

        it('Executed', async function () {
          await this.helper.propose();
          await this.helper.waitForSnapshot();
          await this.helper.vote({ vote: VoteType.For }, { from: voter1 });
          await this.helper.waitForVotingDeadline();
          await this.helper.tally();
          await this.helper.waitForDeadline();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Active);
          await this.helper.execute();
          expect(await this.mock.state(this.proposal.id)).to.be.bignumber.equal(ProposalState.Executed);
        });
      });
    });
  }
});
