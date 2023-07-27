//SPDX-License-Identifier: Apache-2.0

pragma solidity ^0.8.9;

import "../cryptography/Cryptography.sol";
import "../cryptography/BN256Group.sol";

/// @title A basic contract for private voting.
/// @notice This contract implements a very basic logic:
/// - The contract must be re-deployed for each separate voting event.
/// - The starting/stopping of the voting phase is not authomatized, but left to the authority.
/// - There is no voter authentication. Every address can cast a vote (still double voting from
///   the same address is prevented).
/// - There is a single tallying entity in control of the secret key behind the voting event
///   public key. This entity can easily deanonymize individual votes.
/// - Only yes/no votes are possible.
contract Voting is Cryptography, BN256Group {
    enum Status {
        INIT,
        DECLARED,
        VOTING,
        TALLYING,
        FINI
    }

    /// @notice The address authorized to declare the election public key, and start/stop the
    /// voting period. It is set during deployment to the deployer address.
    address authority;
    /// @notice The current status of the voting event.
    Status public status;
    /// @notice The EC-ElGamal public key of the voting event, used for encrypting votes.
    /// The corresponding secret key is known by the authority responsible for tallying.
    GroupElement pk;
    /// @notice A running encrypted tally of the votes cast so far.
    EncryptedVote encryptedTally;
    /// @notice The result (i.e. number of yes) of the voting event. Becomes available after
    /// tallying.
    uint result;
    /// @notice A map to keep track of the address which have already voted to prevent
    /// double voting.
    mapping(address => bool) addressAlreadyVoted;
    /// @notice A map to keep track of the proofs of vote well-formedness already sent, to
    /// prevent possible attacks on ballot privacy.
    mapping(bytes32 => bool) proofHashAlreadySeen;

    event VotingStarted(GroupElement pk);
    event EncryptedVoteCast(address voter);
    event VotingStopped(EncryptedVote tally);
    event Result(uint result);

    error Unauthorized(address caller);
    error WrongStatus(Status status);
    error DoubleVoting(address caller);
    error DoubleProof(address caller);
    error ProofVerificationFailure();

    constructor() {
        authority = msg.sender;
        status = Status.INIT;
    }

    /// @notice Declare the EC-ElGamal public key of the voting event. Can only be called by
    /// authority.
    /// @param pk_ The public key
    /// @param proof The proof of knowledge of the secret key behind pk_. This is generated
    /// off-chain by the cryptographic backend.
    function declarePk(
        GroupElement calldata pk_,
        ProofSkKnowledge calldata proof
    ) external {
        if (status != Status.INIT) {
            revert WrongStatus(status);
        }
        if (msg.sender != authority) {
            revert Unauthorized(msg.sender);
        }
        if (!_verifySkKnowledge(proof, pk_)) {
            revert ProofVerificationFailure();
        }

        pk = pk_;
        status = Status.DECLARED;
    }

    /// @notice Start the voting phase. Can only be called by authority.
    function startVotingPhase() external {
        if (status != Status.DECLARED) {
            revert WrongStatus(status);
        }
        if (msg.sender != authority) {
            revert Unauthorized(msg.sender);
        }

        status = Status.VOTING;

        emit VotingStarted(pk);
    }

    /// @notice Cast an encrypted yes/no vote. Can be called by anyone.
    /// @param vote The vote, encrypted via EC-Elgamal using public key pk
    /// @param proof The proof of well-formedness of vote. This is generated
    /// off-chain by the cryptographic backend.
    function castVote(
        ProofVoteWellFormedness calldata proof,
        EncryptedVote calldata vote
    ) external {
        if (status != Status.VOTING) {
            revert WrongStatus(status);
        }
        if (addressAlreadyVoted[msg.sender]) {
            revert DoubleVoting(msg.sender);
        }
        bytes32 proofHash = keccak256(
            bytes.concat(
                bytes32(Scalar.unwrap(proof.r0)),
                bytes32(Scalar.unwrap(proof.r1)),
                bytes16(Challenge.unwrap(proof.c0)),
                bytes16(Challenge.unwrap(proof.c1))
            )
        );
        if (proofHashAlreadySeen[proofHash]) {
            revert DoubleProof(msg.sender);
        }
        if (!_verifyVoteWellFormedness(proof, vote, pk)) {
            revert ProofVerificationFailure();
        }

        encryptedTally = _addVotes(encryptedTally, vote);
        addressAlreadyVoted[msg.sender] = true;
        proofHashAlreadySeen[proofHash] = true;

        emit EncryptedVoteCast(msg.sender);
    }

    /// @notice Stop the voting phase. Can only be called by authority.
    function stopVotingPhase() external {
        if (status != Status.VOTING) {
            revert WrongStatus(status);
        }
        if (msg.sender != authority) {
            revert Unauthorized(msg.sender);
        }

        status = Status.TALLYING;

        emit VotingStopped(encryptedTally);
    }

    /// @notice Announce the result of tallying. Can be called by anyone, but requires knowledge
    /// of the secret key behind public key pk in order to actually perform the tallying off-chain
    /// and produce a valid proof of correct decryption.
    /// @param decryptedTally The result of decrypting the encrypted tally
    /// @param proof The proof of correct decryption of the encrypted tally. This is generated
    /// off-chain by the cryptographic backend.
    function tally(
        ProofCorrectDecryption calldata proof,
        uint decryptedTally
    ) external {
        if (status != Status.TALLYING) {
            revert WrongStatus(status);
        }
        if (
            !_verifyCorrectDecryption(proof, encryptedTally, decryptedTally, pk)
        ) {
            revert ProofVerificationFailure();
        }

        result = decryptedTally;
        status = Status.FINI;

        emit Result(result);
    }

    /// @notice A getter for the public key of the voting event. Can only be called after the
    /// public key has been declared.
    function getPk() external view returns (GroupElement memory pk_) {
        if (status == Status.INIT) {
            revert WrongStatus(status);
        }

        pk_ = pk;
    }

    /// @notice A getter for the result of the voting event. Can only be called after tallying.
    function getResult() external view returns (uint result_) {
        if (status != Status.FINI) {
            revert WrongStatus(status);
        }

        result_ = result;
    }
}
