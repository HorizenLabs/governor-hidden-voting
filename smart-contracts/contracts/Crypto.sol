//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "hardhat/console.sol";

struct CurvePoint {
    uint256 x;
    uint256 y;
}

type Scalar is uint256;

type Challenge is uint128;

/// @dev The order of bn256.G1 curve. Value taken from variable bn256.Order,
/// go package github.com/ethereum/go-ethereum/crypto/bn256/cloudflare
uint256 constant ORDER = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
/// @dev The prime of bn256.G1 base field. Value taken from variable bn256.P,
/// go package github.com/ethereum/go-ethereum/crypto/bn256/cloudflare
uint256 constant P = 21888242871839275222246405745257275088696311157297823662689037894645226208583;

uint256 constant EC_ADD_ADDRESS = 0x06;
uint256 constant EC_ADD_INPUT_LENGTH = 0x80;
uint256 constant EC_ADD_OUTPUT_LENGTH = 0x40;

uint256 constant EC_MUL_ADDRESS = 0x07;
uint256 constant EC_MUL_INPUT_LENGTH = 0x60;
uint256 constant EC_MUL_OUTPUT_LENGTH = 0x40;

struct EncryptedVote {
    CurvePoint a;
    CurvePoint b;
}

struct EncryptedTally {
    EncryptedVote vote;
    uint numVoters;
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

error ProofVerificationFailure();

/// @dev The generator of bn256.G1 group.
/// Coordinate values taken from go unit test TestCoordinateValues
function g() pure returns (CurvePoint memory) {
    return CurvePoint({x: 1, y: 2});
}

/// @dev Inverse of generator of bn256.G1 group.
/// Coordinate values taken from go unit test TestCoordinateValues
function gNeg() pure returns (CurvePoint memory) {
    return CurvePoint({x: 1, y: P - 2});
}

/// @dev A simple wrapper around call to precompiled smart contract
/// for point addition on elliptic curve bn256.G1
function ecAdd(
    CurvePoint memory a,
    CurvePoint memory b
) view returns (CurvePoint memory result) {
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

/// @dev A simple wrapper around call to precompiled smart contract
/// for scalar multiplication on elliptic curve bn256.G1
function ecMul(
    CurvePoint memory a,
    Scalar s
) view returns (CurvePoint memory result) {
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

function fiatShamirChallenge(bytes memory data) pure returns (Challenge) {
    return Challenge.wrap(uint128(uint256(keccak256(data))));
}

function scalar(Challenge challenge) pure returns (Scalar) {
    return Scalar.wrap(uint256(Challenge.unwrap(challenge)));
}

function neg(Scalar s) pure returns (Scalar) {
    if (Scalar.unwrap(s) == 0) {
        return s;
    }
    return Scalar.wrap(ORDER - Scalar.unwrap(s));
}

function verifyVoteWellFormedness(
    ProofVoteWellFormedness calldata proof,
    EncryptedVote calldata vote,
    CurvePoint storage pk
) view {
    if (Scalar.unwrap(proof.r0) >= ORDER || Scalar.unwrap(proof.r1) >= ORDER) {
        revert ProofVerificationFailure();
    }
    CurvePoint memory a0 = ecAdd(
        ecMul(g(), proof.r0),
        ecMul(vote.a, neg(scalar(proof.c0)))
    );
    CurvePoint memory a1 = ecAdd(
        ecMul(g(), proof.r1),
        ecMul(vote.a, neg(scalar(proof.c1)))
    );
    CurvePoint memory b0 = ecAdd(
        ecMul(pk, proof.r0),
        ecMul(vote.b, neg(scalar(proof.c0)))
    );
    CurvePoint memory b1 = ecAdd(
        ecMul(pk, proof.r1),
        ecMul(ecAdd(vote.b, gNeg()), neg(scalar(proof.c1)))
    );
    Challenge c = fiatShamirChallenge(
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
        if (
            Challenge.unwrap(c) !=
            Challenge.unwrap(proof.c0) + Challenge.unwrap(proof.c1)
        ) {
            revert ProofVerificationFailure();
        }
    }
}

function verifySkKnowledge(
    ProofSkKnowledge calldata proof,
    CurvePoint calldata pk
) view {
    if (Scalar.unwrap(proof.s) >= ORDER) {
        revert ProofVerificationFailure();
    }
    CurvePoint memory v = ecAdd(
        ecMul(g(), proof.s),
        ecMul(pk, neg(scalar(proof.c)))
    );
    Challenge c = fiatShamirChallenge(
        bytes.concat(bytes32(pk.x), bytes32(pk.y), bytes32(v.x), bytes32(v.y))
    );
    if (Challenge.unwrap(c) != Challenge.unwrap(proof.c)) {
        revert ProofVerificationFailure();
    }
}

function verifyCorrectDecryption(
    ProofCorrectDecryption calldata proof,
    EncryptedVote storage encryptedVote,
    uint decryptedVote,
    CurvePoint storage pk
) view {
    if (Scalar.unwrap(proof.s) >= ORDER) {
        revert ProofVerificationFailure();
    }
    CurvePoint memory d = ecAdd(
        encryptedVote.b,
        ecMul(g(), neg(Scalar.wrap(decryptedVote)))
    );
    verifyCorrectDecryptionInternal(proof, encryptedVote, d, pk);
}

function verifyCorrectDecryptionInternal(
    ProofCorrectDecryption calldata proof,
    EncryptedVote storage encryptedVote,
    CurvePoint memory d,
    CurvePoint storage pk
) view {
    CurvePoint memory u = ecAdd(
        ecMul(encryptedVote.a, proof.s),
        ecMul(d, neg(scalar(proof.c)))
    );
    CurvePoint memory v = ecAdd(
        ecMul(g(), proof.s),
        ecMul(pk, neg(scalar(proof.c)))
    );
    Challenge c = fiatShamirChallenge(
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
    if (Challenge.unwrap(c) != Challenge.unwrap(proof.c)) {
        revert ProofVerificationFailure();
    }
}

/// @dev Encrypted tally is initialized to a non-zero value to "warm-up" the storage
/// location, so that the first voter does not incur an additional gas cost compared
/// to subsequent voters.
function initTally(EncryptedTally storage tally) {
    tally.numVoters = 1;
    tally.vote.a = g();
    tally.vote.b = g();
}

/// @dev When voting is finished, the dummy value added during tally initialization
/// must be subtracted.
function finalizeTally(EncryptedTally storage tally) {
    tally.numVoters--;
    tally.vote.a = ecAdd(tally.vote.a, gNeg());
    tally.vote.b = ecAdd(tally.vote.b, gNeg());
}

function updateTally(
    EncryptedTally storage tally,
    EncryptedVote calldata vote
) {
    tally.numVoters++;
    tally.vote.a = ecAdd(tally.vote.a, vote.a);
    tally.vote.b = ecAdd(tally.vote.b, vote.b);
}
