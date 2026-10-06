// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {Script} from "forge-std/Script.sol";

import {ChapterPaymentVault} from "../src/ChapterPaymentVault.sol";
import {MockEthUsdPriceFeed} from "../src/mocks/MockEthUsdPriceFeed.sol";
import {MockUSDC} from "../src/mocks/MockUSDC.sol";

contract DeployLocalChapterPaymentVault is Script {
    uint256 private constant ANVIL_CHAIN_ID = 31_337;
    string private constant ANVIL_MNEMONIC = "test test test test test test test test test test test junk";
    uint32 private constant FUNDED_ANVIL_ACCOUNTS = 5;
    uint256 private constant TEST_USDC_PER_ACCOUNT = 1000e6;
    int256 private constant DEFAULT_ETH_USD_PRICE = 3000e8;
    uint256 private constant MAXIMUM_PRICE_AGE = 1 hours;
    string private constant DEPLOYMENT_FILE = "./deployments/local.json";

    error OnlyForLocalAnvil(uint256 chainId);

    function run() external {
        if (block.chainid != ANVIL_CHAIN_ID) {
            revert OnlyForLocalAnvil(block.chainid);
        }
        uint256 deployerKey = vm.envUint("DEPLOYER_PRIVATE_KEY");
        address deployer = vm.addr(deployerKey);
        address treasury = vm.envOr("TREASURY_ADDRESS", deployer);
        address paymentSigner = vm.envAddress("PAYMENT_SIGNER_ADDRESS");
        int256 ethUsdPrice = vm.envOr("LOCAL_ETH_USD_PRICE", DEFAULT_ETH_USD_PRICE);

        vm.startBroadcast(deployerKey);
        MockUSDC usdc = new MockUSDC();
        MockEthUsdPriceFeed priceFeed = new MockEthUsdPriceFeed(deployer, ethUsdPrice, true);
        ChapterPaymentVault vault =
            new ChapterPaymentVault(deployer, treasury, paymentSigner, usdc, priceFeed, MAXIMUM_PRICE_AGE);
        for (uint32 accountIndex = 0; accountIndex < FUNDED_ANVIL_ACCOUNTS; accountIndex++) {
            usdc.mint(vm.addr(vm.deriveKey(ANVIL_MNEMONIC, accountIndex)), TEST_USDC_PER_ACCOUNT);
        }
        vm.stopBroadcast();

        string memory deploymentKey = "local";
        vm.serializeUint(deploymentKey, "chain_id", block.chainid);
        vm.serializeAddress(deploymentKey, "chapter_payment_vault", address(vault));
        vm.serializeAddress(deploymentKey, "usdc", address(usdc));
        vm.serializeAddress(deploymentKey, "eth_usd_price_feed", address(priceFeed));
        vm.serializeAddress(deploymentKey, "payment_signer", paymentSigner);
        string memory deploymentJson = vm.serializeAddress(deploymentKey, "treasury", treasury);
        vm.writeJson(deploymentJson, DEPLOYMENT_FILE);
    }
}
