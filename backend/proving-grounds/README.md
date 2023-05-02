# Private voting

This service provides access to the cryptographic operations and to the zk-proof generation capabilities required for the private voting protocol. The service can be registered into a `proving-grounds` proxy instance.

# Notation and Data Models

All capabilities expect JSON inputs for each argument and the following specifies the representation for these data models.

- **Number** a signed 64-bit integer.
- **BasePoint** hex encoding of an element of the base field of the bn256.G1 curve.
- **CurvePoint** a point on the bn256.G1 curve
  - x: BasePoint
  - y: BasePoint
- **Scalar** hex encoding of an element of the scalar field of the bn256.G1 curve.
- **Challenge** hex encoding of a 128-bit challenge.
- **KeyPair** an EC-ElGamal key pair on the bn256.G1 curve
  - pk: CurvePoint
  - sk: Scalar
- **EncryptedVote** the EC-ElGamal encryption of a vote
  - a: CurvePoint
  - b: CurvePoint
- **ProofSkKnowledge**
  - s: Scalar
  - c: Challenge
- **ProofVoteWellFormedness**
  - r0: Scalar
  - r1: Scalar
  - c0: Challenge
  - c1: Challenge
- **ProofCorrectDecryption**
  - s: Scalar
  - c: Challenge

# Capabilities

The following capabilities are supported

| Capability                                              | ID                                     |
|---------------------------------------------------------|----------------------------------------|
| [`new_key_pair_with_proof`](#new-key-pair-with-proof)   | `924e949e-0f98-42c6-a4ff-bb2b1c8693e0` |
| [`encrypt_vote_with_proof`](#encrypt-vote-with-proof)   | `772e6272-8c49-463f-b2d1-92088ae06da1` |
| [`decrypt_tally_with_proof`](#decrypt-tally-with-proof) | `c238d864-ae22-4db5-b3d1-83c41cb8b4dd` |

## New Key Pair With Proof

Capability id: `924e949e-0f98-42c6-a4ff-bb2b1c8693e0`

Generate a new EC-ElGamal key pair `(pk, sk)` over the bn256.G1 curve, together with a zk-proof that `sk` is the discrete logarithm of `pk`.

Inputs: None

Outputs:
1. **KeyPair** the key pair
2. **ProofSkKnowledge** the zk-proof

## Encrypt Vote With Proof

Capability id: `772e6272-8c49-463f-b2d1-92088ae06da1`

Encrypt a vote using EC-ElGamal encryption over the bn256.G1 curve, and generate a zk-proof that the encrypted vote is the encryption of a valid (i.e. 0/1) vote.

Inputs:
1. **Number** the vote (must be either 0 or 1)
2. **CurvePoint** the pk for encrypting the vote

Outputs:
1. **EncryptedVote** the encrypted vote
2. **ProofVoteWellFormedness** the zk-proof

## Decrypt Tally With Proof

Capability id: `c238d864-ae22-4db5-b3d1-83c41cb8b4dd`

Decrypt an encrypted tally, and generate a zk-proof that the decryption is correct.

Inputs:
1. **EncryptedVote** the encrypted tally, which results from the homomorphic sum of encrypted 0/1 votes
2. **Number** an upper bound on the result (e.g. the total number of votes)
3. **KeyPair** the EC-ElGamal key pair used for encrypting/decrypting votes

Outputs:
1. **Number** the result of the decryption
2. **ProofCorrectDecryption** the zk-proof
