// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {ERC1155} from "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";
import {Test} from "forge-std/Test.sol";

contract SetupProbeToken is ERC1155 {
    constructor() ERC1155("https://metadata.test/{id}.json") {}
}

contract ProjectSetupTest is Test {
    function test_OpenZeppelinContractsCompileAndDeploy() public {
        SetupProbeToken probeToken = new SetupProbeToken();
        assertEq(probeToken.uri(1), "https://metadata.test/{id}.json");
    }
}
