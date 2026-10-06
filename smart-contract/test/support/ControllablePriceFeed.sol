// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";

contract ControllablePriceFeed is AggregatorV3Interface {
    uint8 public immutable decimals;
    int256 private latestAnswer;
    uint256 private latestStartedAt;
    uint256 private latestUpdatedAt;
    uint80 private latestRoundId;

    constructor(uint8 feedDecimals, int256 initialAnswer) {
        decimals = feedDecimals;
        setAnswer(initialAnswer, block.timestamp, block.timestamp);
    }

    function setAnswer(int256 newAnswer, uint256 startedAt, uint256 updatedAt) public {
        latestRoundId += 1;
        latestAnswer = newAnswer;
        latestStartedAt = startedAt;
        latestUpdatedAt = updatedAt;
    }

    function description() external pure returns (string memory) {
        return "controllable test feed";
    }

    function version() external pure returns (uint256) {
        return 1;
    }

    function getRoundData(uint80) external view returns (uint80, int256, uint256, uint256, uint80) {
        return (latestRoundId, latestAnswer, latestStartedAt, latestUpdatedAt, latestRoundId);
    }

    function latestRoundData() external view returns (uint80, int256, uint256, uint256, uint80) {
        return (latestRoundId, latestAnswer, latestStartedAt, latestUpdatedAt, latestRoundId);
    }
}
