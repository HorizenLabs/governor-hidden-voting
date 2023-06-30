//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "./IGroup.sol";

abstract contract Cryptography is IGroup {
    type Challenge is uint128;

    struct EncryptedVote {
        GroupElement a;
        GroupElement b;
    }

    struct ProofSkKnowledge {
        Scalar s;
        Challenge c;
    }

    struct ProofCorrectDecryption {
        Scalar s;
        Challenge c;
    }

    struct ProofVoteWellFormedness {
        Scalar r0;
        Scalar r1;
        Challenge c0;
        Challenge c1;
    }

    function _fiatShamirChallenge(
        bytes memory data
    ) internal pure returns (Challenge) {
        return Challenge.wrap(uint128(uint256(keccak256(data))));
    }

    function _scalar(Challenge challenge) internal pure returns (Scalar) {
        return Scalar.wrap(uint256(Challenge.unwrap(challenge)));
    }

    function _verifyVoteWellFormedness(
        ProofVoteWellFormedness memory proof,
        EncryptedVote memory vote,
        GroupElement memory pk
    ) internal view virtual returns (bool) {
        if (
            Scalar.unwrap(proof.r0) >= order() ||
            Scalar.unwrap(proof.r1) >= order()
        ) {
            return false;
        }
        GroupElement memory a0 = groupAdd(
            scalarMul(generator(), proof.r0),
            scalarMul(vote.a, scalarNeg(_scalar(proof.c0)))
        );
        GroupElement memory a1 = groupAdd(
            scalarMul(generator(), proof.r1),
            scalarMul(vote.a, scalarNeg(_scalar(proof.c1)))
        );
        GroupElement memory b0 = groupAdd(
            scalarMul(pk, proof.r0),
            scalarMul(vote.b, scalarNeg(_scalar(proof.c0)))
        );
        GroupElement memory b1 = groupAdd(
            scalarMul(pk, proof.r1),
            scalarMul(
                groupAdd(vote.b, groupNeg(generator())),
                scalarNeg(_scalar(proof.c1))
            )
        );
        Challenge c = _fiatShamirChallenge(
            bytes.concat(
                bytes32(pk.x),
                bytes32(pk.y),
                bytes32(vote.a.x),
                bytes32(vote.a.y),
                bytes32(vote.b.x),
                bytes32(vote.b.y),
                bytes32(a0.x),
                bytes32(a0.y),
                bytes32(b0.x),
                bytes32(b0.y),
                bytes32(a1.x),
                bytes32(a1.y),
                bytes32(b1.x),
                bytes32(b1.y)
            )
        );
        // wrapping this into an unchecked block, since challenges use mod 2^128 arithmetics
        unchecked {
            return (Challenge.unwrap(c) ==
                Challenge.unwrap(proof.c0) + Challenge.unwrap(proof.c1));
        }
    }

    function _verifySkKnowledge(
        ProofSkKnowledge calldata proof,
        GroupElement calldata pk
    ) internal view virtual returns (bool) {
        if (Scalar.unwrap(proof.s) >= order()) {
            return false;
        }
        GroupElement memory v = groupAdd(
            scalarMul(generator(), proof.s),
            scalarMul(pk, scalarNeg(_scalar(proof.c)))
        );
        Challenge c = _fiatShamirChallenge(
            bytes.concat(
                bytes32(pk.x),
                bytes32(pk.y),
                bytes32(v.x),
                bytes32(v.y)
            )
        );
        return Challenge.unwrap(c) == Challenge.unwrap(proof.c);
    }

    function _verifyCorrectDecryption(
        ProofCorrectDecryption calldata proof,
        EncryptedVote memory encryptedVote,
        uint decryptedVote,
        GroupElement memory pk
    ) internal view virtual returns (bool) {
        if (Scalar.unwrap(proof.s) >= order()) {
            return false;
        }
        GroupElement memory d = groupAdd(
            encryptedVote.b,
            scalarMul(generator(), scalarNeg(Scalar.wrap(decryptedVote)))
        );
        return verifyCorrectDecryptionInternal(proof, encryptedVote, d, pk);
    }

    function verifyCorrectDecryptionInternal(
        ProofCorrectDecryption calldata proof,
        EncryptedVote memory encryptedVote,
        GroupElement memory d,
        GroupElement memory pk
    ) internal view returns (bool) {
        GroupElement memory u = groupAdd(
            scalarMul(encryptedVote.a, proof.s),
            scalarMul(d, scalarNeg(_scalar(proof.c)))
        );
        GroupElement memory v = groupAdd(
            scalarMul(generator(), proof.s),
            scalarMul(pk, scalarNeg(_scalar(proof.c)))
        );
        Challenge c = _fiatShamirChallenge(
            bytes.concat(
                bytes32(pk.x),
                bytes32(pk.y),
                bytes32(encryptedVote.a.x),
                bytes32(encryptedVote.a.y),
                bytes32(encryptedVote.b.x),
                bytes32(encryptedVote.b.y),
                bytes32(u.x),
                bytes32(u.y),
                bytes32(v.x),
                bytes32(v.y)
            )
        );
        return Challenge.unwrap(c) == Challenge.unwrap(proof.c);
    }

    function _addVotes(
        EncryptedVote memory vote0,
        EncryptedVote memory vote1
    ) internal view virtual returns (EncryptedVote memory) {
        return
            EncryptedVote({
                a: groupAdd(vote0.a, vote1.a),
                b: groupAdd(vote0.b, vote1.b)
            });
    }

    function _scaleVote(
        EncryptedVote memory vote,
        Scalar s
    ) internal view virtual returns (EncryptedVote memory) {
        return
            EncryptedVote({a: scalarMul(vote.a, s), b: scalarMul(vote.b, s)});
    }
}
