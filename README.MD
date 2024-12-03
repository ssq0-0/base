# BASE AIO

The software is a multi-module programme designed for automated operations in the BASE blockchain ecosystem. Modules can be combined and configured individually, allowing flexible and personalized workflows.

---

## Features

- **Uniswap:** Automatic decentralized exchange operations.
- **PancakeSwap:** Automatic decentralized exchange operations.
- **Woofi:** Automatic decentralized exchange operations.
- **Dmail:** Generate and send decentralized mails.
- **Zora:** Automatic NFT minting on Zora.
- **NFT2ME:** Automated NFT minting for NFT2ME contracts.
- **Aave:** Interaction with liquidity pools on Aave.
- **Moonwell:** Liquidity pool management on Moonwell.
- **BaseNames:** Domain minting for BASE blockchain.
- **Collector:** Token collection and aggregation.
- **Refuel:** (Coming soon) Refueling ETH across networks.
- **Memory:** Tracks and adapts operations dynamically.

*(More features coming soon!)*

---

## Installation

### Requirements

- **Go** (Version 1.22.2 or newer)
- Git (for cloning the repository)
- Optional: `make` for simplified build and run commands.

### Steps

1. Clone the repository:
```bash
git clone https://github.com/ssq0-0/base.git
cd base
go mod download
go build -o base ./app/main.go   
```
2. Run the application:

```bash
./base
```

3. Or use make(if installed):
```bash
make run
```

## Configuration Guide

The configuration file (`config.json`) is essential for customizing the behavior of the software. Below is a detailed explanation of its fields and their usage.

---

### Wallets (`wallets`)

This section defines the wallets used by the software. Each wallet is described by the following fields:

- **`private_key`**: The private key of your wallet. Required for signing transactions.
- **`base_name`**: (Optional) If you need to mint a domain on BASE, specify the domain name here. For the cheapest options, use names with 10 or more characters.
- **`used_range`**: The percentage of your balance to be used for swaps. For example, with a balance of $100 and `used_range` set to 70, $70 will be used for each swap. It is not recommended to set this value above 90%.
- **`used_range_in_pools`**: Similar to `used_range`, but for liquidity pools (Aave & Moonwell). Recommended value is no more than 30% when combining multiple modules.
- **`bridge`**: If bridging is required, specify the target network (e.g., `optimism`, `arbitrum`). Ensure your account has sufficient ETH in the BASE network to cover gas fees.
- **`token`**: The token to be bridged (e.g., `usdt`, `usdc`). Note: Bridging uses Stargate, and the resulting token in the BASE network will be `usdbc`.
- **`action_num_min` / `action_num_max`**: Minimum and maximum number of actions to be performed. A random number within this range will be chosen.
- **`action_time_window_min` / `action_time_window_max`**: Minimum and maximum delay between actions, in minutes.

---

### Modules (`modules`)

Modules define which functionalities of the software are enabled. Set a module to `true` to enable it, or `false` to disable it.

- Avoid enabling all modules simultaneously.
- The `collector_mod` should always be used separately from other modules.

Example:
```json
"modules": {
  "uniswap": true,
  "pancake": false,
  "woofi": false,
  "dmail": false,
  "zora": true,
  "nft2me": false,
  "aave": true,
  "moonwell": false,
  "basenames": true,
  "collector_mod": false
}
```

## NFT Contracts (`nft_ca`)

The `nft_ca` section defines which NFT contracts will be used for minting operations. Each entry specifies the contract address and the price for minting.

---

### Fields

- **`nf2me`**: 
  - Specify contracts for NFT2ME.
  - Include the price for minting each NFT. For free NFTs, specify `0.0002` to cover gas fees.

- **`zora`**: 
  - Specify contracts for Zora.
  - Provide the minting price in Sparks (Zora's internal currency). Note that the software does not currently support a Sparks-to-ETH calculator, so the price must be calculated manually.

---

### Example Configuration

```json
"nft_ca": {
  "nf2me": {
    "0xContractAddress1": "0.0002",
    "0xContractAddress2": "0.0002"
  },
  "zora": {
    "0xContractAddress3": "0.000051",
    "0xContractAddress4": "0.000042"
  }
}
```

### Important Notes

1. **Contract Updates**:
   - After each session, update the contract list to prevent duplication.
   - The software automatically removes contracts from the list once they are assigned to an account for minting.

2. **Multi-Account Safety**:
   - When using multiple accounts, contracts are assigned uniquely per account during the same session, ensuring no overlaps.

3. **NFT2ME Pricing**:
   - For free NFTs, always specify `0.0002` as the price to cover gas fees.

4. **Zora Pricing**:
   - Zora uses Sparks as its currency. Calculate the equivalent ETH price manually before adding it to the configuration. A calculator for Sparks conversion will be available in future updates.

---

### Example Configuration

Here is a complete example of the `config.json` file, incorporating all wallet, module, and NFT contract settings:

```json
{
  "wallets": [
    {
      "private_key": "your_private_key_here",
      "base_name": "example-domain",
      "used_range": 70,
      "used_range_in_pools": 25,
      "bridge": "optimism",
      "token": "usdt",
      "action_num_min": 15,
      "action_num_max": 20,
      "action_time_window_min": 1,
      "action_time_window_max": 5
    }
  ],
  "modules": {
    "uniswap": true,
    "pancake": false,
    "woofi": false,
    "dmail": false,
    "zora": true,
    "nft2me": true,
    "aave": true,
    "moonwell": false,
    "basenames": true,
    "collector_mod": false
  },
  "nft_ca": {
    "nf2me": {
      "0xContractAddress1": "0.0002",
      "0xContractAddress2": "0.0002"
    },
    "zora": {
      "0xContractAddress3": "0.000051",
      "0xContractAddress4": "0.000042"
    }
  }
}
```

### For additional assistance or troubleshooting, refer to the official documentation or reach out via [support channel](https://t.me/cheifssq).