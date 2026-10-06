// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {Test} from "forge-std/Test.sol";

import {ChapterPaymentVault} from "../src/ChapterPaymentVault.sol";
import {CircleUsdcMinting} from "./support/CircleUsdcMinting.sol";
import {VaultDeployment} from "./support/VaultDeployment.sol";

contract ChapterPaymentVaultForkTest is Test {
    uint256 private constant CHAPTER_PRICE_CENTS = 499;
    uint256 private constant CHAPTER_PRICE_USDC_UNITS = 4_990_000;

    ChapterPaymentVault private vault;
    AggregatorV3Interface private chainlinkEthUsdFeed;
    IERC20 private usdc;
    address private player = makeAddr("player");
    uint256 private paymentSignerKey;
    bool private isForkAvailable;

    function setUp() public {
        string memory rpcUrl = vm.envOr("BASE_SEPOLIA_RPC_URL", string(""));
        if (bytes(rpcUrl).length == 0) {
            return;
        }
        isForkAvailable = true;
        vm.createSelectFork(rpcUrl);
        string memory networkConfig = vm.readFile("./config/networks/base-sepolia.json");
        chainlinkEthUsdFeed = AggregatorV3Interface(vm.parseJsonAddress(networkConfig, ".eth_usd_price_feed"));
        usdc = IERC20(vm.parseJsonAddress(networkConfig, ".payment_token"));
        address paymentSigner;
        (paymentSigner, paymentSignerKey) = makeAddrAndKey("paymentSigner");
        vault = VaultDeployment.deployBehindProxy(
            VaultDeployment.settingsFor(
                address(this),
                makeAddr("treasury"),
                paymentSigner,
                IERC20Metadata(address(usdc)),
                chainlinkEthUsdFeed,
                1 hours
            )
        );
        vm.deal(player, 1 ether);
    }

    modifier onlyWithFork() {
        if (!isForkAvailable) {
            vm.skip(true);
        }
        _;
    }

    function _authorize(bytes32 paymentReference)
        private
        view
        returns (ChapterPaymentVault.PaymentAuthorization memory)
    {
        uint256 deadline = block.timestamp + 30 minutes;
        bytes32 digest = vault.hashPaymentAuthorization(paymentReference, player, CHAPTER_PRICE_CENTS, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(paymentSignerKey, digest);
        return ChapterPaymentVault.PaymentAuthorization({
            paymentReference: paymentReference,
            usdCents: CHAPTER_PRICE_CENTS,
            deadline: deadline,
            signature: abi.encodePacked(r, s, v)
        });
    }

    function test_QuotesEthFromTheLiveChainlinkFeed() public onlyWithFork {
        (, int256 ethUsdPrice,,,) = chainlinkEthUsdFeed.latestRoundData();
        uint256 expectedWei = Math.mulDiv(
            CHAPTER_PRICE_CENTS,
            1e18 * 10 ** chainlinkEthUsdFeed.decimals(),
            SafeCast.toUint256(ethUsdPrice) * 100,
            Math.Rounding.Ceil
        );
        assertEq(vault.quoteWei(CHAPTER_PRICE_CENTS), expectedWei);
    }

    function test_PaysWithEthAtTheLiveChainlinkPrice() public onlyWithFork {
        uint256 requiredWei = vault.quoteWei(CHAPTER_PRICE_CENTS);
        ChapterPaymentVault.PaymentAuthorization memory authorization = _authorize(keccak256("fork-eth-payment"));
        vm.prank(player);
        vault.payWithEth{value: requiredWei * 2}(authorization);
        assertEq(address(vault).balance, requiredWei);
        assertEq(player.balance, 1 ether - requiredWei);
    }

    function test_PaysWithCircleUsdc() public onlyWithFork {
        CircleUsdcMinting circleUsdc = CircleUsdcMinting(address(usdc));
        address masterMinter = circleUsdc.masterMinter();
        vm.startPrank(masterMinter);
        circleUsdc.configureMinter(masterMinter, CHAPTER_PRICE_USDC_UNITS);
        circleUsdc.mint(player, CHAPTER_PRICE_USDC_UNITS);
        vm.stopPrank();

        ChapterPaymentVault.PaymentAuthorization memory authorization = _authorize(keccak256("fork-usdc-payment"));
        vm.startPrank(player);
        usdc.approve(address(vault), CHAPTER_PRICE_USDC_UNITS);
        vault.payWithToken(authorization);
        vm.stopPrank();
        assertEq(usdc.balanceOf(address(vault)), CHAPTER_PRICE_USDC_UNITS);
        assertEq(usdc.balanceOf(player), 0);
    }
}
