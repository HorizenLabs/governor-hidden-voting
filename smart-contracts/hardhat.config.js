require("@nomicfoundation/hardhat-toolbox");
require('@nomiclabs/hardhat-truffle5');
require("hardhat-gas-reporter");
require('hardhat-exposed');

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.18",
    settings: {
      optimizer: {
        enabled: true,
        runs: 1000,
      },
    },
  },
  gasReporter: {
    enabled: (process.env.REPORT_GAS) ? true : false
  },
  exposed: {
    initializers: true,
  },
  networks: {
    hardhat: {
      allowUnlimitedContractSize: true,
    },
  },
};
