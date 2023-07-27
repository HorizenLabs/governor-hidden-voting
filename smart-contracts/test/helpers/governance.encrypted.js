const { GovernorHelper } = require('../../lib/openzeppelin-contracts/test/helpers/governance.js');
const { forward } = require('../../lib/openzeppelin-contracts/test/helpers/time.js');

function concatOpts(args, opts = null) {
    return opts ? args.concat(opts) : args;
}

function stringifyPk(pk) {
    return "x:" + pk.x + ",y:" + pk.y;
}

VoteType = {
    Against: 0,
    For: 1,
};

class GovernorEncryptedHelper extends GovernorHelper {
    constructor(governor) {
        super(governor);
        this.keyStore = new Map();
    }

    async waitForVotingDeadline(offset = 0) {
        const proposal = this.currentProposal;
        const timepoint = await this.governor.votingDeadline(proposal.id)
        return forward[this.mode](timepoint.addn(offset));
    }

    async vote(vote = {}, opts = null) {
        const proposal = this.currentProposal;

        var encryptedVote, proof;
        if (vote.vote == VoteType.Against || vote.vote == VoteType.For) {
            const pk = vote.pk ? vote.pk : await this.governor.getPk(proposal.id);
            ({ encryptedVote, proof } = await goEncryptVoteWithProof(vote.vote, pk));
        } else {
            const { keyPair } = await goNewKeyPairWithProof();
            ({ encryptedVote, proof } = await goEncryptVoteWithProof(VoteType.Against, keyPair.pk));
        }
        return this.governor.castEncryptedVote(...concatOpts([proposal.id, encryptedVote, proof], opts));
    }

    async addKeyPair() {
        const { keyPair, proof } = await goNewKeyPairWithProof();
        const key = stringifyPk(keyPair.pk);
        this.keyStore.set(key, { sk: keyPair.sk, proof });
        return keyPair.pk;
    }

    async initialize(args = {}, opts = null) {
        const key = stringifyPk({ x: args.pk.x, y: args.pk.y });
        const proof = this.keyStore.get(key).proof;
        return this.governor.initialize(...concatOpts([args.pk, proof], opts));
    }

    async updateCurrentPk(args = {}, opts = null) {
        const key = stringifyPk({ x: args.pk.x, y: args.pk.y });
        const proof = this.keyStore.get(key).proof;
        return this.governor.updateCurrentPk(...concatOpts([args.pk, proof], opts));
    }

    async tally(args = {}, opts = null) {
        const proposal = this.currentProposal;

        var result, proof;
        if (args.fake) {
            const { keyPair } = await goNewKeyPairWithProof();
            const { encryptedVote } = await goEncryptVoteWithProof(VoteType.Against, keyPair.pk);
            ({ result, proof } = await goDecryptTallyWithProof(encryptedVote, 1, keyPair));
        } else {
            const encryptedTally = await this.governor.getTally(proposal.id);
            const castVotes = await this.governor.getCastVotes(proposal.id);
            const pk = await this.governor.getPk(proposal.id);
            const key = stringifyPk({ x: pk.x, y: pk.y });
            const sk = this.keyStore.get(key).sk;
            const keyPair = { pk, sk };
            ({ result, proof } = await goDecryptTallyWithProof(...concatOpts([encryptedTally, castVotes.toNumber(), keyPair], opts)));
        }

        return this.governor.tally(proposal.id, proof, result);
    }
}

module.exports = {
    GovernorEncryptedHelper, VoteType, stringifyPk
};
