// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {Script} from "forge-std/Script.sol";

import {ChapterPaymentVault} from "../src/ChapterPaymentVault.sol";

contract UpgradeChapterPaymentVault is Script {
    string private constant DEFAULT_DEPLOYMENT_NAME = "local";

    function run() external {
        uint256 ownerKey = vm.envUint("DEPLOYER_PRIVATE_KEY");
        string memory deploymentPath =
            string.concat("./deployments/", vm.envOr("DEPLOYMENT_NAME", DEFAULT_DEPLOYMENT_NAME), ".json");
        ChapterPaymentVault vault =
            ChapterPaymentVault(payable(vm.parseJsonAddress(vm.readFile(deploymentPath), ".chapter_payment_vault")));

        vm.startBroadcast(ownerKey);
        ChapterPaymentVault nextImplementation = new ChapterPaymentVault();
        vault.upgradeToAndCall(address(nextImplementation), "");
        vm.stopBroadcast();

        vm.writeJson(vm.toString(address(nextImplementation)), deploymentPath, ".implementation");
    }
}
