//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "./IGroup.sol";

contract BN256Group is IGroup {
    /// @dev The order of bn256.G1 curve. Value taken from variable bn256.Order,
    /// go package github.com/ethereum/go-ethereum/crypto/bn256/cloudflare
    uint256 constant ORDER =
        21888242871839275222246405745257275088548364400416034343698204186575808495617;
    /// @dev The prime of bn256.G1 base field. Value taken from variable bn256.P,
    /// go package github.com/ethereum/go-ethereum/crypto/bn256/cloudflare
    uint256 constant P =
        21888242871839275222246405745257275088696311157297823662689037894645226208583;

    uint256 constant EC_ADD_ADDRESS = 0x06;
    uint256 constant EC_ADD_INPUT_LENGTH = 0x80;
    uint256 constant EC_ADD_OUTPUT_LENGTH = 0x40;

    uint256 constant EC_MUL_ADDRESS = 0x07;
    uint256 constant EC_MUL_INPUT_LENGTH = 0x60;
    uint256 constant EC_MUL_OUTPUT_LENGTH = 0x40;

    function generator()
        internal
        view
        virtual
        override
        returns (GroupElement memory)
    {
        return GroupElement({x: 1, y: 2});
    }

    function order() public view virtual override returns (uint256) {
        return ORDER;
    }

    function groupAdd(
        GroupElement memory a,
        GroupElement memory b
    ) internal view virtual override returns (GroupElement memory result) {
        bytes32[4] memory input;
        input[0] = bytes32(a.x);
        input[1] = bytes32(a.y);
        input[2] = bytes32(b.x);
        input[3] = bytes32(b.y);
        assembly {
            let success := staticcall(
                gas(),
                EC_ADD_ADDRESS,
                input,
                EC_ADD_INPUT_LENGTH,
                result,
                EC_ADD_OUTPUT_LENGTH
            )
            switch success
            case 0 {
                revert(0, 0)
            }
        }
    }

    function groupNeg(
        GroupElement memory a
    ) internal view virtual override returns (GroupElement memory) {
        return GroupElement({x: a.x, y: P - a.y});
    }

    function scalarMul(
        GroupElement memory a,
        Scalar s
    ) internal view virtual override returns (GroupElement memory result) {
        bytes32[3] memory input;
        input[0] = bytes32(a.x);
        input[1] = bytes32(a.y);
        input[2] = bytes32(Scalar.unwrap(s));
        assembly {
            let success := staticcall(
                gas(),
                EC_MUL_ADDRESS,
                input,
                EC_MUL_INPUT_LENGTH,
                result,
                EC_MUL_OUTPUT_LENGTH
            )
            switch success
            case 0 {
                revert(0, 0)
            }
        }
    }
}
