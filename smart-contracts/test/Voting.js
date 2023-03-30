const { expect } = require("chai");
const { loadFixture } = require("@nomicfoundation/hardhat-network-helpers");

const Status = {
    INIT: 0,
    DECLARED: 1,
    VOTING: 2,
    TALLYING: 3,
    FINI: 4
}

describe("Voting contract", function () {
    async function deployedFixture() {
        const Voting = await ethers.getContractFactory("Voting");
        const addr = await ethers.getSigners();
        const owner = addr[0];

        const hardhatVoting = await Voting.deploy();
        await hardhatVoting.deployed();

        const data = require("./test_data.json");
        return { Voting, hardhatVoting, owner, addr, data };
    }

    async function pkDeclaredFixture() {
        const { hardhatVoting, owner, addr, data } = await loadFixture(
            deployedFixture
        );
        await hardhatVoting.declarePk(data.PkA, data.ProofSkKnowledgeA);
        return { hardhatVoting, owner, addr, data };
    }

    async function votingStartedFixture() {
        const { hardhatVoting, owner, addr, data } = await loadFixture(
            pkDeclaredFixture
        );
        await hardhatVoting.startVotingPhase();
        return { hardhatVoting, owner, addr, data };
    }

    async function votesCastFixture() {
        const { hardhatVoting, owner, addr, data } = await loadFixture(
            votingStartedFixture
        );
        for (let i = 0; i < data.EncryptedVotesValid.length; i++) {
            await hardhatVoting.connect(addr[i]).castVote(
                data.ProofsVoteWellFormednessValid[i],
                data.EncryptedVotesValid[i]
            );
        }
        return { hardhatVoting, owner, addr, data };
    }

    async function votingStoppedFixture() {
        const { hardhatVoting, owner, addr, data } = await loadFixture(
            votesCastFixture
        );
        await hardhatVoting.stopVotingPhase();
        return { hardhatVoting, owner, addr, data };
    }

    describe("declarePk", function () {
        it("Should succeed", async function () {
            const { hardhatVoting, data } = await loadFixture(
                deployedFixture
            );
            await hardhatVoting.declarePk(data.PkA, data.ProofSkKnowledgeA);
        });

        it("Should set pk to correct value", async function () {
            const { hardhatVoting, data } = await loadFixture(
                deployedFixture
            );
            await hardhatVoting.declarePk(data.PkA, data.ProofSkKnowledgeA);
            const [pkx, pky] = await hardhatVoting.getPk();
            expect(pkx).to.equal(ethers.BigNumber.from(data.PkA.x));
            expect(pky).to.equal(ethers.BigNumber.from(data.PkA.y));
        });

        it("Should revert if invoked twice", async function () {
            const { hardhatVoting, data } = await loadFixture(
                deployedFixture
            );
            await hardhatVoting.declarePk(data.PkA, data.ProofSkKnowledgeA);
            await expect(hardhatVoting.declarePk(data.PkB, data.ProofSkKnowledgeB))
                .to.be.revertedWithCustomError(hardhatVoting, "WrongStatus")
                .withArgs(Status.DECLARED);
        });

        it("Should revert if invoked by unauthorized user", async function () {
            const { hardhatVoting, data, addr } = await loadFixture(
                deployedFixture
            );
            await expect(hardhatVoting.connect(addr[1]).declarePk(data.PkA, data.ProofSkKnowledgeA))
                .to.be.revertedWithCustomError(hardhatVoting, "Unauthorized")
                .withArgs(await addr[1].getAddress());
        });

        it("Should revert if proof is invalid", async function () {
            const { hardhatVoting, data } = await loadFixture(
                deployedFixture
            );
            await expect(hardhatVoting.declarePk(data.PkA, data.ProofSkKnowledgeB))
                .to.be.revertedWithCustomError(hardhatVoting, "ProofVerificationFailure")
        });

        it("Should revert if pk is not well-formed", async function () {
            const { hardhatVoting, data } = await loadFixture(
                deployedFixture
            );
            let pk = JSON.parse(JSON.stringify(data.PkA));
            // swap x and y coordinates to create an invalid pk
            [pk.x, pk.y] = [pk.y, pk.x]
            await expect(hardhatVoting.declarePk(pk, data.ProofSkKnowledgeA))
                .to.be.revertedWithoutReason;
        });
    });

    describe("startVotingPhase", function () {
        it("Should succeed", async function () {
            const { hardhatVoting } = await loadFixture(
                pkDeclaredFixture
            );
            await hardhatVoting.startVotingPhase();
        });

        it("Should revert if invoked twice", async function () {
            const { hardhatVoting } = await loadFixture(
                pkDeclaredFixture
            );
            await hardhatVoting.startVotingPhase();
            await expect(hardhatVoting.startVotingPhase())
                .to.be.revertedWithCustomError(hardhatVoting, "WrongStatus")
                .withArgs(Status.VOTING);
        });

        it("Should revert if invoked by unauthorized user", async function () {
            const { hardhatVoting, addr } = await loadFixture(
                pkDeclaredFixture
            );
            await expect(hardhatVoting.connect(addr[1]).startVotingPhase())
                .to.be.revertedWithCustomError(hardhatVoting, "Unauthorized")
                .withArgs(await addr[1].getAddress());
        });
    });

    describe("castVote", function () {
        it("Should succeed", async function () {
            const { hardhatVoting, addr, data } = await loadFixture(
                votingStartedFixture
            );
            for (let i = 0; i < data.EncryptedVotesValid.length; i++) {
                await hardhatVoting.connect(addr[i]).castVote(
                    data.ProofsVoteWellFormednessValid[i],
                    data.EncryptedVotesValid[i]
                );
            }
        });

        it("Should prevent double voting", async function () {
            const { hardhatVoting, addr, data } = await loadFixture(
                votingStartedFixture
            );
            hardhatVoting.connect(addr[1]).castVote(
                data.ProofsVoteWellFormednessValid[1],
                data.EncryptedVotesValid[1]
            );
            await expect(
                hardhatVoting.connect(addr[1]).castVote(
                    data.ProofsVoteWellFormednessValid[2],
                    data.EncryptedVotesValid[2])
            ).to.be.revertedWithCustomError(hardhatVoting, "DoubleVoting");
        });

        it("Should prevent proof reuse", async function () {
            const { hardhatVoting, addr, data } = await loadFixture(
                votingStartedFixture
            );
            hardhatVoting.connect(addr[1]).castVote(
                data.ProofsVoteWellFormednessValid[1],
                data.EncryptedVotesValid[1]
            );
            await expect(
                hardhatVoting.connect(addr[2]).castVote(
                    data.ProofsVoteWellFormednessValid[1],
                    data.EncryptedVotesValid[1])
            ).to.be.revertedWithCustomError(hardhatVoting, "DoubleProof");
        });

        it("Should revert if proof is invalid (different vote)", async function () {
            const { hardhatVoting, addr, data } = await loadFixture(
                votingStartedFixture
            );
            await expect(
                hardhatVoting.castVote(
                    data.ProofsVoteWellFormednessValid[1],
                    data.EncryptedVotesValid[2])
            ).to.be.revertedWithCustomError(hardhatVoting, "ProofVerificationFailure");
        });

        it("Should revert if proof is invalid (different pk)", async function () {
            const { hardhatVoting, addr, data } = await loadFixture(
                votingStartedFixture
            );
            for (let i = 0; i < data.EncryptedVotesInvalid.length; i++) {
                await expect(hardhatVoting.connect(addr[i]).castVote(
                    data.ProofsVoteWellFormednessInvalid[i],
                    data.EncryptedVotesInvalid[i]
                )).to.be.revertedWithCustomError(hardhatVoting, "ProofVerificationFailure");
            }
        });
    });

    describe("stopVotingPhase", function () {
        it("Should succeed", async function () {
            const { hardhatVoting } = await loadFixture(
                votesCastFixture
            );
            await hardhatVoting.stopVotingPhase();
        });

        it("Should revert if invoked twice", async function () {
            const { hardhatVoting } = await loadFixture(
                votesCastFixture
            );
            await hardhatVoting.stopVotingPhase();
            await expect(hardhatVoting.stopVotingPhase())
                .to.be.revertedWithCustomError(hardhatVoting, "WrongStatus")
                .withArgs(Status.TALLYING);
        });

        it("Should revert if invoked by unauthorized user", async function () {
            const { hardhatVoting, addr } = await loadFixture(
                votesCastFixture
            );
            await expect(hardhatVoting.connect(addr[1]).stopVotingPhase())
                .to.be.revertedWithCustomError(hardhatVoting, "Unauthorized")
                .withArgs(await addr[1].getAddress());
        });
    });

    describe("tally", function () {
        it("Should succeed", async function () {
            const { hardhatVoting, data } = await loadFixture(
                votingStoppedFixture
            );
            await hardhatVoting.tally(data.ProofCorrectDecryptionValid, data.Result);
        });

        it("Should set result to correct value", async function () {
            const { hardhatVoting, data } = await loadFixture(
                votingStoppedFixture
            );
            await hardhatVoting.tally(data.ProofCorrectDecryptionValid, data.Result);
            expect(await hardhatVoting.getResult()).to.equal(data.Result);
        });

        it("Should revert if invoked twice", async function () {
            const { hardhatVoting, data } = await loadFixture(
                votingStoppedFixture
            );
            await hardhatVoting.tally(data.ProofCorrectDecryptionValid, data.Result);
            await expect(hardhatVoting.tally(data.ProofCorrectDecryptionValid, data.Result))
                .to.be.revertedWithCustomError(hardhatVoting, "WrongStatus")
                .withArgs(Status.FINI);
        });

        it("Should revert if proof is invalid", async function () {
            const { hardhatVoting, data } = await loadFixture(
                votingStoppedFixture
            );
            await expect(hardhatVoting.tally(data.ProofCorrectDecryptionInvalid, data.Result))
                .to.be.revertedWithCustomError(hardhatVoting, "ProofVerificationFailure")
        });

        it("Should revert if result is wrong", async function () {
            const { hardhatVoting, data } = await loadFixture(
                votingStoppedFixture
            );
            await expect(hardhatVoting.tally(data.ProofCorrectDecryptionValid, data.Result + 1))
                .to.be.revertedWithCustomError(hardhatVoting, "ProofVerificationFailure")
        });
    });
});
