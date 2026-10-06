// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

interface CircleUsdcMinting {
    function masterMinter() external view returns (address);

    function configureMinter(address minter, uint256 minterAllowedAmount) external returns (bool);

    function mint(address recipient, uint256 amount) external returns (bool);
}
