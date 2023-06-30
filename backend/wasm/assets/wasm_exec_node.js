function loadEVotingBackend() {
	globalThis.require = require;
	globalThis.fs = require("fs");
	globalThis.path = require("path")
	globalThis.TextEncoder = require("util").TextEncoder;
	globalThis.TextDecoder = require("util").TextDecoder;

	globalThis.performance = {
		now() {
			const [sec, nsec] = process.hrtime();
			return sec * 1000 + nsec / 1000000;
		},
	};

	const crypto = require("crypto");
	globalThis.crypto = {
		getRandomValues(b) {
			crypto.randomFillSync(b);
		},
	};

	require("./wasm_exec");

	const go = new Go();

	const wasmPath = path.join(__dirname, 'main.wasm');
	return WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject)
		.then((result) => { go.run(result.instance); })
		.then(_ => {
			return new Promise(function (resolve, reject) {
				let timeout = setTimeout(() => reject("failed to load wasm functions"), 5000);
				let retry = setInterval(function () {
					if (
						typeof (goNewKeyPairWithProof) !== "function" ||
						typeof (goEncryptVoteWithProof) !== "function" ||
						typeof (goDecryptTallyWithProof) !== "function" ||
						typeof (goAddEncryptedVotes) !== "function"
					) {
						return;
					}
					clearInterval(retry);
					clearTimeout(timeout);
					resolve("loaded wasm functions");
				}, 100);
			});
		});
}

module.exports = { loadEVotingBackend };
