// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";

import {ChapterPaymentVault} from "../../src/ChapterPaymentVault.sol";

library VaultDeployment {
    function deployBehindProxy(ChapterPaymentVault.InitializationSettings memory settings)
        internal
        returns (ChapterPaymentVault)
    {
        ChapterPaymentVault implementation = new ChapterPaymentVault();
        ERC1967Proxy proxy =
            new ERC1967Proxy(address(implementation), abi.encodeCall(ChapterPaymentVault.initialize, (settings)));
        return ChapterPaymentVault(payable(address(proxy)));
    }

    function settingsFor(
        address owner,
        address treasury,
        address paymentSigner,
        IERC20Metadata paymentToken,
        AggregatorV3Interface ethUsdPriceFeed,
        uint256 maximumPriceAge
    ) internal pure returns (ChapterPaymentVault.InitializationSettings memory) {
        return ChapterPaymentVault.InitializationSettings({
            owner: owner,
            treasury: treasury,
            paymentSigner: paymentSigner,
            paymentToken: paymentToken,
            ethUsdPriceFeed: ethUsdPriceFeed,
            sequencerUptimeFeed: AggregatorV3Interface(address(0)),
            maximumPriceAge: maximumPriceAge
        });
    }
}
