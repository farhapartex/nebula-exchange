// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

import {IEthUsdPriceFeed} from "../interfaces/IEthUsdPriceFeed.sol";

contract MockEthUsdPriceFeed is IEthUsdPriceFeed, Ownable {
    uint8 private constant PRICE_DECIMALS = 8;

    bool public immutable isAlwaysFresh;
    int256 private latestPrice;
    uint256 private latestUpdatedAt;
    uint80 private latestRoundId;

    constructor(address initialOwner, int256 initialPrice, bool reportsAlwaysFresh) Ownable(initialOwner) {
        isAlwaysFresh = reportsAlwaysFresh;
        _recordPrice(initialPrice, block.timestamp);
    }

    function setPrice(int256 newPrice) external onlyOwner {
        _recordPrice(newPrice, block.timestamp);
    }

    function setPriceUpdatedAt(int256 newPrice, uint256 updatedAt) external onlyOwner {
        _recordPrice(newPrice, updatedAt);
    }

    function decimals() external pure returns (uint8) {
        return PRICE_DECIMALS;
    }

    function latestRoundData() external view returns (uint80, int256, uint256, uint256, uint80) {
        uint256 reportedUpdatedAt = isAlwaysFresh ? block.timestamp : latestUpdatedAt;
        return (latestRoundId, latestPrice, reportedUpdatedAt, reportedUpdatedAt, latestRoundId);
    }

    function _recordPrice(int256 newPrice, uint256 updatedAt) private {
        latestRoundId += 1;
        latestPrice = newPrice;
        latestUpdatedAt = updatedAt;
    }
}
