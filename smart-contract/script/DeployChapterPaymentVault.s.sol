// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {Script} from "forge-std/Script.sol";

import {ChapterPaymentVault} from "../src/ChapterPaymentVault.sol";

contract DeployChapterPaymentVault is Script {
    string private constant DEFAULT_NETWORK_CONFIG = "./config/networks/base-sepolia.json";
    string private constant DEFAULT_DEPLOYMENT_NAME = "local";
    uint256 private constant DEFAULT_MAXIMUM_PRICE_AGE = 1 hours;

    function run() external {
        uint256 deployerKey = vm.envUint("DEPLOYER_PRIVATE_KEY");
        address deployer = vm.addr(deployerKey);
        string memory networkConfig = vm.readFile(vm.envOr("NETWORK_CONFIG", DEFAULT_NETWORK_CONFIG));
        ChapterPaymentVault.InitializationSettings memory settings = ChapterPaymentVault.InitializationSettings({
            owner: deployer,
            treasury: vm.envOr("TREASURY_ADDRESS", deployer),
            paymentSigner: vm.envAddress("PAYMENT_SIGNER_ADDRESS"),
            paymentToken: IERC20Metadata(vm.parseJsonAddress(networkConfig, ".payment_token")),
            ethUsdPriceFeed: AggregatorV3Interface(vm.parseJsonAddress(networkConfig, ".eth_usd_price_feed")),
            sequencerUptimeFeed: AggregatorV3Interface(vm.parseJsonAddress(networkConfig, ".sequencer_uptime_feed")),
            maximumPriceAge: vm.envOr("MAXIMUM_PRICE_AGE", DEFAULT_MAXIMUM_PRICE_AGE)
        });

        vm.startBroadcast(deployerKey);
        ChapterPaymentVault implementation = new ChapterPaymentVault();
        ERC1967Proxy proxy =
            new ERC1967Proxy(address(implementation), abi.encodeCall(ChapterPaymentVault.initialize, (settings)));
        vm.stopBroadcast();

        string memory deploymentKey = "deployment";
        vm.serializeUint(deploymentKey, "chain_id", block.chainid);
        vm.serializeAddress(deploymentKey, "chapter_payment_vault", address(proxy));
        vm.serializeAddress(deploymentKey, "implementation", address(implementation));
        vm.serializeAddress(deploymentKey, "payment_token", address(settings.paymentToken));
        vm.serializeAddress(deploymentKey, "eth_usd_price_feed", address(settings.ethUsdPriceFeed));
        vm.serializeAddress(deploymentKey, "payment_signer", settings.paymentSigner);
        string memory deploymentJson = vm.serializeAddress(deploymentKey, "treasury", settings.treasury);
        vm.writeJson(deploymentJson, deploymentFile());
    }

    function deploymentFile() private view returns (string memory) {
        return string.concat("./deployments/", vm.envOr("DEPLOYMENT_NAME", DEFAULT_DEPLOYMENT_NAME), ".json");
    }
}
