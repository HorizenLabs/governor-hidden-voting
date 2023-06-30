This module allows to compile the Go backend in wasm, so that it can be easily called from javascript. In particular, the following functions are exposed to javascript:
- `goNewKeyPairWithProof`
- `goEncryptVoteWithProof`
- `goDecryptTallyWithProof`
- `goAddEncryptedVotes`

To compile, run the command `make`. This will compile the files inside `cmd/wasm` and place the resulting `main.wasm` file inside the `assets` directory.

The file `assets/wasm_exec.js` is copied from the Go distribution, and performs the necessary setup to call wasm files compiled from Go.

The script `assets/index.html` allows to invoke the backend functionalities from a browser.

The function `loadEVotingBackend()`, exported by module `assets/wasm_exec_node.js`, instead, must be invoked in a Node.js environment before being able to invoke the backend functionalities.

`cmd/server` is a very basic server for serving the contents of the directory `assets`, in order to test calling the backend from a browser.