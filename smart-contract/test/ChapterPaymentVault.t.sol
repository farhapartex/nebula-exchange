// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IERC20Errors} from "@openzeppelin/contracts/interfaces/draft-IERC6093.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {Test} from "forge-std/Test.sol";

import {ChapterPaymentVault} from "../src/ChapterPaymentVault.sol";
import {MockEthUsdPriceFeed} from "../src/mocks/MockEthUsdPriceFeed.sol";
import {MockUSDC} from "../src/mocks/MockUSDC.sol";

contract ChapterPaymentVaultTest is Test {
    int256 private constant ETH_PRICE_3000_USD = 3000e8;
    uint256 private constant CHAPTER_PRICE_CENTS = 499;
    uint256 private constant ONE_HOUR = 1 hours;
    uint256 private constant CHECKOUT_LIFETIME = 30 minutes;

    ChapterPaymentVault private vault;
    MockUSDC private usdc;
    MockEthUsdPriceFeed private priceFeed;

    address private owner = makeAddr("owner");
    address private treasury = makeAddr("treasury");
    address private player = makeAddr("player");
    address private otherPlayer = makeAddr("otherPlayer");
    uint256 private paymentSignerKey;
    address private paymentSigner;

    event PaymentReceived(
        bytes32 indexed paymentReference,
        address indexed payer,
        address indexed asset,
        uint256 amountPaid,
        uint256 usdCents
    );

    function setUp() public {
        vm.warp(1_790_000_000);
        (paymentSigner, paymentSignerKey) = makeAddrAndKey("paymentSigner");
        usdc = new MockUSDC();
        priceFeed = new MockEthUsdPriceFeed(owner, ETH_PRICE_3000_USD, false);
        vault = new ChapterPaymentVault(owner, treasury, paymentSigner, usdc, priceFeed, ONE_HOUR);
        vm.deal(player, 10 ether);
        vm.deal(otherPlayer, 10 ether);
        usdc.mint(player, 1000e6);
    }

    function _authorize(bytes32 paymentReference, address payer, uint256 usdCents)
        private
        view
        returns (ChapterPaymentVault.PaymentAuthorization memory)
    {
        return
            _authorizeWithKey(paymentSignerKey, paymentReference, payer, usdCents, block.timestamp + CHECKOUT_LIFETIME);
    }

    function _authorizeWithKey(
        uint256 signerKey,
        bytes32 paymentReference,
        address payer,
        uint256 usdCents,
        uint256 deadline
    ) private view returns (ChapterPaymentVault.PaymentAuthorization memory) {
        bytes32 digest = vault.hashPaymentAuthorization(paymentReference, payer, usdCents, deadline);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerKey, digest);
        return ChapterPaymentVault.PaymentAuthorization({
            paymentReference: paymentReference,
            usdCents: usdCents,
            deadline: deadline,
            signature: abi.encodePacked(r, s, v)
        });
    }

    function test_PayWithEthChargesTheFeedPriceAndRefundsTheRest() public {
        bytes32 paymentReference = keccak256("payment-1");
        uint256 requiredWei = vault.quoteWei(CHAPTER_PRICE_CENTS);
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(paymentReference, player, CHAPTER_PRICE_CENTS);
        assertEq(requiredWei, 1_663_333_333_333_334);

        vm.expectEmit(address(vault));
        emit PaymentReceived(paymentReference, player, address(0), requiredWei, CHAPTER_PRICE_CENTS);
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);

        assertEq(address(vault).balance, requiredWei);
        assertEq(player.balance, 10 ether - requiredWei);
        assertTrue(vault.isPaymentReferencePaid(paymentReference));
    }

    function test_PayWithEthRejectsTooLittleEth() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        uint256 requiredWei = vault.quoteWei(CHAPTER_PRICE_CENTS);
        vm.expectRevert(
            abi.encodeWithSelector(ChapterPaymentVault.InsufficientEthSent.selector, requiredWei, requiredWei - 1)
        );
        vm.prank(player);
        vault.payWithEth{value: requiredWei - 1}(authorization);
    }

    function test_PayWithTokenTakesTheExactUsdcAmount() public {
        bytes32 paymentReference = keccak256("payment-usdc");
        vm.startPrank(player);
        usdc.approve(address(vault), 4_990_000);
        vm.expectEmit(address(vault));
        emit PaymentReceived(paymentReference, player, address(usdc), 4_990_000, CHAPTER_PRICE_CENTS);
        vault.payWithToken(_authorize(paymentReference, player, CHAPTER_PRICE_CENTS));
        vm.stopPrank();

        assertEq(usdc.balanceOf(address(vault)), 4_990_000);
        assertEq(usdc.balanceOf(player), 1000e6 - 4_990_000);
    }

    function test_PayWithTokenNeedsAnApproval() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-usdc"), player, CHAPTER_PRICE_CENTS);
        vm.expectRevert(
            abi.encodeWithSelector(IERC20Errors.ERC20InsufficientAllowance.selector, address(vault), 0, 4_990_000)
        );
        vm.prank(player);
        vault.payWithToken(authorization);
    }

    function test_APaymentReferenceCanOnlyBePaidOnce() public {
        bytes32 paymentReference = keccak256("payment-1");
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(paymentReference, player, CHAPTER_PRICE_CENTS);
        vm.startPrank(player);
        vault.payWithEth{value: 1 ether}(authorization);
        vm.expectRevert(abi.encodeWithSelector(ChapterPaymentVault.PaymentAlreadyMade.selector, paymentReference));
        vault.payWithEth{value: 1 ether}(authorization);
        vm.stopPrank();
    }

    function test_AnotherWalletCannotUseAPlayersAuthorization() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        vm.expectRevert(ChapterPaymentVault.InvalidPaymentAuthorization.selector);
        vm.prank(otherPlayer);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_ALowerPriceThanTheSignedOneIsRejected() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        authorization.usdCents = 1;
        vm.expectRevert(ChapterPaymentVault.InvalidPaymentAuthorization.selector);
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_OnlyThePaymentSignerCanAuthorize() public {
        (, uint256 strangerKey) = makeAddrAndKey("stranger");
        ChapterPaymentVault.PaymentAuthorization memory authorization = _authorizeWithKey(
            strangerKey, keccak256("payment-1"), player, CHAPTER_PRICE_CENTS, block.timestamp + CHECKOUT_LIFETIME
        );
        vm.expectRevert(ChapterPaymentVault.InvalidPaymentAuthorization.selector);
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_AnExpiredAuthorizationIsRejected() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        vm.warp(authorization.deadline + 1);
        vm.expectRevert(
            abi.encodeWithSelector(ChapterPaymentVault.PaymentAuthorizationExpired.selector, authorization.deadline)
        );
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_AStaleOrBrokenPriceStopsEthPayments() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        uint256 staleUpdatedAt = block.timestamp - ONE_HOUR - 1;
        vm.prank(owner);
        priceFeed.setPriceUpdatedAt(ETH_PRICE_3000_USD, staleUpdatedAt);
        vm.expectRevert(abi.encodeWithSelector(ChapterPaymentVault.StaleEthUsdPrice.selector, staleUpdatedAt));
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);

        vm.prank(owner);
        priceFeed.setPrice(0);
        vm.expectRevert(abi.encodeWithSelector(ChapterPaymentVault.InvalidEthUsdPrice.selector, int256(0)));
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_PausedVaultTakesNoPayments() public {
        ChapterPaymentVault.PaymentAuthorization memory authorization =
            _authorize(keccak256("payment-1"), player, CHAPTER_PRICE_CENTS);
        vm.prank(owner);
        vault.pause();
        vm.expectRevert(Pausable.EnforcedPause.selector);
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);

        vm.prank(owner);
        vault.unpause();
        vm.prank(player);
        vault.payWithEth{value: 1 ether}(authorization);
    }

    function test_WithdrawalsGoOnlyToTheTreasury() public {
        vm.startPrank(player);
        vault.payWithEth{value: 1 ether}(_authorize(keccak256("payment-eth"), player, CHAPTER_PRICE_CENTS));
        usdc.approve(address(vault), 4_990_000);
        vault.payWithToken(_authorize(keccak256("payment-usdc"), player, CHAPTER_PRICE_CENTS));
        vm.stopPrank();
        uint256 vaultEth = address(vault).balance;

        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, player));
        vm.prank(player);
        vault.withdrawEth(vaultEth);

        vm.startPrank(owner);
        vault.withdrawEth(vaultEth);
        vault.withdrawToken(IERC20(address(usdc)), 4_990_000);
        vm.stopPrank();
        assertEq(treasury.balance, vaultEth);
        assertEq(usdc.balanceOf(treasury), 4_990_000);
        assertEq(address(vault).balance, 0);
    }

    function test_OwnerControlsCannotBeMisused() public {
        vm.startPrank(owner);
        vm.expectRevert(ChapterPaymentVault.ZeroAddress.selector);
        vault.setTreasury(address(0));
        vm.expectRevert(abi.encodeWithSelector(ChapterPaymentVault.InvalidMaximumPriceAge.selector, 2 days));
        vault.setMaximumPriceAge(2 days);
        vm.expectRevert(ChapterPaymentVault.OwnershipCannotBeRenounced.selector);
        vault.renounceOwnership();
        vm.stopPrank();

        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, player));
        vm.prank(player);
        vault.setPaymentSigner(player);
    }

    function test_PlainEthTransfersAreRefused() public {
        vm.prank(player);
        (bool isAccepted,) = address(vault).call{value: 1 ether}("");
        assertFalse(isAccepted);
    }

    function testFuzz_QuotedWeiNeverUndercharges(uint256 usdCents, uint256 ethPrice) public {
        usdCents = bound(usdCents, 1, 10_000_000);
        ethPrice = bound(ethPrice, 1e8, 1_000_000e8);
        vm.prank(owner);
        priceFeed.setPrice(SafeCast.toInt256(ethPrice));

        uint256 quotedWei = vault.quoteWei(usdCents);
        assertGe(quotedWei * ethPrice * 100, usdCents * 1e18 * 1e8);
        assertLt((quotedWei - 1) * ethPrice * 100, usdCents * 1e18 * 1e8);
    }
}
