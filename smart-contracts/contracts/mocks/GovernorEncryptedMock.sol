//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "../cryptography/BN256Group.sol";
import "../openzeppelin-voting/GovernorEncrypted.sol";
import "../openzeppelin-voting/UpdateablePublicKey.sol";
import "../openzeppelin-voting/GovernorEncryptedSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorProposalThreshold.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorSettings.sol";
import "@openzeppelin/contracts/governance/extensions/GovernorVotesQuorumFraction.sol";

abstract contract GovernorEncryptedMock is
    BN256Group,
    Cryptography,
    GovernorProposalThreshold,
    GovernorVotesQuorumFraction,
    GovernorEncrypted,
    GovernorSettings,
    GovernorEncryptedSettings,
    UpdateablePublicKey
{
    function propose(
        address[] memory targets,
        uint256[] memory values,
        bytes[] memory calldatas,
        string memory description
    )
        public
        virtual
        override(Governor, GovernorEncrypted, GovernorProposalThreshold)
        returns (uint256 proposalId)
    {
        return super.propose(targets, values, calldatas, description);
    }

    function proposalThreshold()
        public
        view
        virtual
        override(Governor, GovernorSettings)
        returns (uint256)
    {
        return super.proposalThreshold();
    }

    function castVote(
        uint256 proposalId,
        uint8 support
    ) public virtual override(Governor, GovernorEncrypted) returns (uint256) {
        return super.castVote(proposalId, support);
    }

    function castVoteWithReason(
        uint256 proposalId,
        uint8 support,
        string calldata reason
    ) public virtual override(Governor, GovernorEncrypted) returns (uint256) {
        return super.castVoteWithReason(proposalId, support, reason);
    }

    function castVoteWithReasonAndParams(
        uint256 proposalId,
        uint8 support,
        string calldata reason,
        bytes memory params
    ) public virtual override(Governor, GovernorEncrypted) returns (uint256) {
        return
            super.castVoteWithReasonAndParams(
                proposalId,
                support,
                reason,
                params
            );
    }

    function castVoteBySig(
        uint256 proposalId,
        uint8 support,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public virtual override(Governor, GovernorEncrypted) returns (uint256) {
        return super.castVoteBySig(proposalId, support, v, r, s);
    }

    function castVoteWithReasonAndParamsBySig(
        uint256 proposalId,
        uint8 support,
        string calldata reason,
        bytes memory params,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) public virtual override(Governor, GovernorEncrypted) returns (uint256) {
        return
            super.castVoteWithReasonAndParamsBySig(
                proposalId,
                support,
                reason,
                params,
                v,
                r,
                s
            );
    }
}
