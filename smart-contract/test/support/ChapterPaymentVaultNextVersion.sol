// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {ChapterPaymentVault} from "../../src/ChapterPaymentVault.sol";

contract ChapterPaymentVaultNextVersion is ChapterPaymentVault {
    function implementationVersion() external pure override returns (uint256) {
        return 2;
    }
}
