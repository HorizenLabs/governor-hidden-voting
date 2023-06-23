//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "@openzeppelin/contracts/governance/utils/IVotes.sol";

contract DiscretizedVotes is IVotes {
    IVotes private immutable _token;
    uint256 public minWeight;

    constructor(IVotes token, uint256 _minWeight) {
        _token = token;
        minWeight = _minWeight;
    }

    function getVotes(
        address account
    ) external view override returns (uint256) {
        return _token.getVotes(account) / minWeight;
    }

    function getPastVotes(
        address account,
        uint256 timepoint
    ) external view override returns (uint256) {
        return _token.getPastVotes(account, timepoint) / minWeight;
    }

    function getPastTotalSupply(
        uint256 timepoint
    ) external view override returns (uint256) {
        return _token.getPastTotalSupply(timepoint) / minWeight;
    }

    function delegates(
        address account
    ) external view override returns (address) {
        return _token.delegates(account);
    }

    function delegate(address delegatee) external override {
        return _token.delegate(delegatee);
    }

    function delegateBySig(
        address delegatee,
        uint256 nonce,
        uint256 expiry,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external override {
        return _token.delegateBySig(delegatee, nonce, expiry, v, r, s);
    }
}
