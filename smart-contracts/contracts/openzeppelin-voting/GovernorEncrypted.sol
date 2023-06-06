//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "../cryptography/Cryptography.sol";
import "@openzeppelin/contracts/governance/Governor.sol";
import "@openzeppelin/contracts/utils/math/SafeCast.sol";

/**
 * @dev Extension of {Governor} for counting votes encrypted via EC El-Gamal.
 */
abstract contract GovernorEncrypted is Cryptography, Governor {
    using SafeCast for uint256;

    struct ProposalVote {
        EncryptedVote tally;
        GroupElement pk;
        uint256 forVotes;
        uint256 castVotes;
        uint64 votingEnd;
        mapping(address => bool) hasVoted;
    }

    mapping(uint256 => ProposalVote) private _proposalVotes;

    /**
     * @dev The duration of the tallying period.
     */
    function tallyingPeriod() public view virtual returns (uint256);

    /**
     * @dev The deadline for voting on proposal identified by proposalId.
     */
    function votingDeadline(
        uint256 proposalId
    ) public view virtual returns (uint64) {
        return _proposalVotes[proposalId].votingEnd;
    }

    /**
     * @dev A getter for the current election public key.
     */
    function _getCurrentPk() internal virtual returns (GroupElement storage);

    /**
     * @dev See {IGovernor-COUNTING_MODE}.
     */
    // solhint-disable-next-line func-name-mixedcase
    function COUNTING_MODE()
        public
        pure
        virtual
        override
        returns (string memory)
    {
        return "support=ignore&quorum=bravo&params=bn256helios";
    }

    /**
     * @dev See {IGovernor-hasVoted}.
     */
    function hasVoted(
        uint256 proposalId,
        address account
    ) public view virtual override returns (bool) {
        ProposalVote storage proposal = _proposalVotes[proposalId];

        return proposal.hasVoted[account];
    }

    /**
     * @dev See {IGovernor-propose}.
     */
    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    ) public virtual override returns (uint256) {
        GroupElement storage currentPk = _getCurrentPk();

        uint256 proposalId = super.propose(
            targets,
            values,
            calldatas,
            description
        );
        ProposalVote storage proposal = _proposalVotes[proposalId];

        uint256 tallyingDeadline = proposalDeadline(proposalId) -
            tallyingPeriod();

        proposal.votingEnd = SafeCast.toUint64(tallyingDeadline);
        proposal.pk = currentPk;

        return proposalId;
    }

    /**
     * @dev See {Governor-_quorumReached}.
     */
    function _quorumReached(
        uint256 proposalId
    ) internal view virtual override returns (bool) {
        ProposalVote storage proposalVote = _proposalVotes[proposalId];

        return quorum(proposalSnapshot(proposalId)) <= proposalVote.forVotes;
    }

    /**
     * @dev See {Governor-_voteSucceeded}.
     */
    function _voteSucceeded(
        uint256 proposalId
    ) internal view virtual override returns (bool) {
        ProposalVote storage proposalVote = _proposalVotes[proposalId];

        return
            proposalVote.forVotes >
            proposalVote.castVotes - proposalVote.forVotes;
    }

    /**
     * @dev See {Governor-_countVote}.
     * In this module, support is ignored.
     * Instead, params must contain the serialization of the encrypted vote, and of the vote well-formedness proof.
     */
    function _countVote(
        uint256 proposalId,
        address account,
        uint8, //support
        uint256 weight,
        bytes memory params
    ) internal virtual override {
        ProposalVote storage proposalVote = _proposalVotes[proposalId];

        require(
            params.length == 256,
            "GovernorCountingEncrypted: wrong params length"
        );

        require(
            !proposalVote.hasVoted[account],
            "GovernorCountingEncrypted: vote already cast"
        );
        proposalVote.hasVoted[account] = true;

        ProofVoteWellFormedness memory proof;
        EncryptedVote memory encryptedVote;
        (encryptedVote, proof) = abi.decode(
            params,
            (EncryptedVote, ProofVoteWellFormedness)
        );
        require(
            _verifyVoteWellFormedness(
                proof,
                encryptedVote,
                _proposalVotes[proposalId].pk
            ),
            "GovernorCountingEncrypted: proof verification failed"
        );

        EncryptedVote memory scaledVote = _scaleVote(
            encryptedVote,
            Scalar.wrap(weight)
        );
        proposalVote.tally = _addVotes(proposalVote.tally, scaledVote);
        proposalVote.castVotes += weight;
    }

    /**
     * @dev See {IGovernor-castVote}.
     */
    function castVote(
        uint256, //proposalId
        uint8 //support
    ) public virtual override returns (uint256) {
        revert("GovernorCountingEncrypted: castVote unavailable");
    }

    /**
     * @dev See {IGovernor-castVoteWithReason}.
     */
    function castVoteWithReason(
        uint256, //proposalId
        uint8, //support
        string calldata //reason
    ) public virtual override returns (uint256) {
        revert("GovernorCountingEncrypted: castVoteWithReason unavailable");
    }

    /**
     * @dev See {IGovernor-castVoteWithReasonAndParams}.
     */
    function castVoteWithReasonAndParams(
        uint256, //proposalId
        uint8, //support
        string calldata, //reason
        bytes memory //params
    ) public virtual override returns (uint256) {
        revert(
            "GovernorCountingEncrypted: castVoteWithReasonAndParams unavailable"
        );
    }

    /**
     * @dev See {IGovernor-castVoteBySig}.
     */
    function castVoteBySig(
        uint256, //proposalId
        uint8, //support
        uint8, //v
        bytes32, //r
        bytes32 //s
    ) public virtual override returns (uint256) {
        revert("GovernorCountingEncrypted: castVoteBySig unavailable");
    }

    /**
     * @dev See {IGovernor-castVoteWithReasonAndParamsBySig}.
     */
    function castVoteWithReasonAndParamsBySig(
        uint256, //proposalId
        uint8, //support
        string calldata, //reason
        bytes memory, //params
        uint8, //v
        bytes32, //r
        bytes32 //s
    ) public virtual override returns (uint256) {
        revert(
            "GovernorCountingEncrypted: castVoteWithReasonAndParamsBySig unavailable"
        );
    }

    /**
     * @dev Wrapper around {Governor-_castVote} to cast an encrypted vote
     */
    function castEncryptedVote(
        uint256 proposalId,
        EncryptedVote calldata vote,
        ProofVoteWellFormedness calldata proof
    ) public virtual returns (uint256) {
        address voter = _msgSender();
        uint8 support = 0;
        string memory reason = "";
        bytes memory params = abi.encode(vote, proof);
        uint256 weight = _castVote(proposalId, voter, support, reason, params);

        ProposalVote storage proposalVote = _proposalVotes[proposalId];
        uint256 currentTimepoint = clock();
        require(
            currentTimepoint <= proposalVote.votingEnd,
            "GovernorCountingEncrypted: voting is over"
        );
        return weight;
    }

    /**
     * @dev Function to be used by tallying authority to post the result of tallying.
     * Must be invoked after votingDeadline() and before proposalDeadline().
     */
    function tally(
        uint256 proposalId,
        ProofCorrectDecryption calldata proof,
        uint forVotes
    ) public {
        ProposalVote storage proposalVote = _proposalVotes[proposalId];
        uint256 currentTimepoint = clock();
        ProposalState status = state(proposalId);
        require(
            status == ProposalState.Active &&
                currentTimepoint > proposalVote.votingEnd,
            "GovernorCountingEncrypted: tallying not currently active"
        );
        require(
            _verifyCorrectDecryption(
                proof,
                proposalVote.tally,
                forVotes,
                proposalVote.pk
            ),
            "GovernorCountingEncrypted: proof verification failed"
        );
        proposalVote.forVotes = forVotes;
    }

    /**
     * @dev Accessor to proposal public key.
     */
    function getPk(
        uint256 proposalId
    ) public view returns (GroupElement memory) {
        return _proposalVotes[proposalId].pk;
    }

    /**
     * @dev Accessor to total weight currently cast on the proposal.
     */
    function getCastVotes(uint256 proposalId) public view returns (uint256) {
        return _proposalVotes[proposalId].castVotes;
    }

    /**
     * @dev Accessor to current encrypted tally of the proposal.
     */
    function getTally(
        uint256 proposalId
    ) public view returns (EncryptedVote memory) {
        return _proposalVotes[proposalId].tally;
    }
}
