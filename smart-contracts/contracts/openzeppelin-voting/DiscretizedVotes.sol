//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "@openzeppelin/contracts/governance/utils/IVotes.sol";

/**
 * @dev Discretize the voting power of an {IVotes} instance using a decorator pattern.
 * Voting power and total supply are divided by the quantity minWeight, which is declared
 * at deployment. In this way voting power is always a multiple of minWeight. This allows
 * using ERC-20 tokens for private voting: if using an ERC-20 token directly, with full
 * resolution of voting power, then tallying could become very inefficient, due to the need
 * to solve a discrete log with a big upper bound (i.e. the total weight cast).
 */
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
