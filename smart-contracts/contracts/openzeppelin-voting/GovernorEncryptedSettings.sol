//SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.9;

import "./GovernorEncrypted.sol";

/**
 * @dev Extension of {GovernorCountingEncrypted} for settings updatable through governance.
 */
abstract contract GovernorEncryptedSettings is GovernorEncrypted {
    uint256 private _tallyingPeriod;

    event TallyingPeriodSet(uint256 oldVotingPeriod, uint256 newVotingPeriod);

    /**
     * @dev Initialize the governance parameters.
     */
    constructor(uint256 initialTallyingPeriod) {
        _setTallyingPeriod(initialTallyingPeriod);
    }

    /**
     * @dev See {GovernorCountingEncrypted-votingPeriod}.
     */
    function tallyingPeriod() public view virtual override returns (uint256) {
        return _tallyingPeriod;
    }

    /**
     * @dev Update the tallying period. This operation can only be performed through a governance proposal.
     *
     * Emits a {TallyingPeriodSet} event.
     */
    function setTallyingPeriod(
        uint256 newTallyingPeriod
    ) public virtual onlyGovernance {
        _setTallyingPeriod(newTallyingPeriod);
    }

    /**
     * @dev Internal setter for the tallying period.
     *
     * Emits a {TallyingPeriodSet} event.
     */
    function _setTallyingPeriod(uint256 newTallyingPeriod) internal virtual {
        require(
            newTallyingPeriod > 0,
            "GovernorEncryptedSettings: tallying period too low"
        );
        require(
            newTallyingPeriod < votingPeriod(),
            "GovernorEncryptedSettings: tallying period too high"
        );
        emit TallyingPeriodSet(_tallyingPeriod, newTallyingPeriod);
        _tallyingPeriod = newTallyingPeriod;
    }
}
