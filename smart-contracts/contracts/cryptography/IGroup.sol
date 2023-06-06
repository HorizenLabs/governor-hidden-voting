//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

abstract contract IGroup {
    struct GroupElement {
        uint256 x;
        uint256 y;
    }

    type Scalar is uint256;

    function generator() internal view virtual returns (GroupElement memory);

    function order() public view virtual returns (uint256);

    function groupAdd(
        GroupElement memory a,
        GroupElement memory b
    ) internal view virtual returns (GroupElement memory result);

    function scalarMul(
        GroupElement memory a,
        Scalar s
    ) internal view virtual returns (GroupElement memory result);

    function groupNeg(
        GroupElement memory a
    ) internal view virtual returns (GroupElement memory);

    function scalarNeg(Scalar s) internal view virtual returns (Scalar) {
        if (Scalar.unwrap(s) == 0) {
            return s;
        }
        return Scalar.wrap(order() - Scalar.unwrap(s));
    }
}
