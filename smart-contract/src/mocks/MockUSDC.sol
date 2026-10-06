// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract MockUSDC is ERC20 {
    uint8 private constant USDC_DECIMALS = 6;

    constructor() ERC20("Mock USD Coin", "USDC") {}

    function decimals() public pure override returns (uint8) {
        return USDC_DECIMALS;
    }

    function mint(address recipient, uint256 amount) external {
        _mint(recipient, amount);
    }
}
