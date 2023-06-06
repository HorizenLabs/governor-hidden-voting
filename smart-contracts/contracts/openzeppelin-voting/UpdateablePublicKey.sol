//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";
import "./GovernorEncrypted.sol";

abstract contract UpdateablePublicKey is Ownable, GovernorEncrypted {
    GroupElement private _currentPk;
    bool private _isInitialized;

    function initialize(
        GroupElement calldata pk,
        ProofSkKnowledge calldata proof
    ) public onlyOwner {
        require(
            !_isInitialized,
            "UpdateablePublicKey: contract already initialized"
        );
        _isInitialized = true;
        _updateCurrentPk(pk, proof);
    }

    function _getCurrentPk()
        internal
        virtual
        override
        returns (GroupElement storage)
    {
        require(
            _isInitialized,
            "UpdateablePublicKey: contract should be initialized"
        );
        return _currentPk;
    }

    function updateCurrentPk(
        GroupElement calldata pk,
        ProofSkKnowledge calldata proof
    ) public virtual onlyGovernance {
        _updateCurrentPk(pk, proof);
    }

    function _updateCurrentPk(
        GroupElement calldata pk,
        ProofSkKnowledge calldata proof
    ) private {
        require(
            _verifySkKnowledge(proof, pk),
            "UpdateablePublicKey: proof verification failed"
        );
        _currentPk = pk;
    }
}
