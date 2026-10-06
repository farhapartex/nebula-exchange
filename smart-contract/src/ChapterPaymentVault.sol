// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Ownable2Step} from "@openzeppelin/contracts/access/Ownable2Step.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Address} from "@openzeppelin/contracts/utils/Address.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuardTransient} from "@openzeppelin/contracts/utils/ReentrancyGuardTransient.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

import {IEthUsdPriceFeed} from "./interfaces/IEthUsdPriceFeed.sol";

contract ChapterPaymentVault is Ownable2Step, Pausable, ReentrancyGuardTransient, EIP712 {
    using SafeERC20 for IERC20;

    struct PaymentAuthorization {
        bytes32 paymentReference;
        uint256 usdCents;
        uint256 deadline;
        bytes signature;
    }

    bytes32 public constant CHAPTER_PAYMENT_TYPEHASH =
        keccak256("ChapterPayment(bytes32 paymentReference,address payer,uint256 usdCents,uint256 deadline)");
    uint256 public constant MAXIMUM_PRICE_AGE_LIMIT = 1 days;
    uint256 private constant CENTS_PER_DOLLAR = 100;
    uint256 private constant WEI_PER_ETHER = 1e18;
    uint8 private constant CENT_DECIMALS = 2;

    IERC20 public immutable paymentToken;
    IEthUsdPriceFeed public immutable ethUsdPriceFeed;
    uint256 private immutable tokenUnitsPerCent;
    uint256 private immutable priceFeedScale;

    address public paymentSigner;
    address public treasury;
    uint256 public maximumPriceAge;
    mapping(bytes32 paymentReference => bool isPaid) public isPaymentReferencePaid;

    event PaymentReceived(
        bytes32 indexed paymentReference,
        address indexed payer,
        address indexed asset,
        uint256 amountPaid,
        uint256 usdCents
    );
    event PaymentSignerChanged(address indexed previousPaymentSigner, address indexed newPaymentSigner);
    event TreasuryChanged(address indexed previousTreasury, address indexed newTreasury);
    event MaximumPriceAgeChanged(uint256 previousMaximumPriceAge, uint256 newMaximumPriceAge);
    event EthWithdrawn(address indexed treasury, uint256 amount);
    event TokenWithdrawn(address indexed token, address indexed treasury, uint256 amount);

    error ZeroAddress();
    error ZeroPaymentAmount();
    error PaymentAlreadyMade(bytes32 paymentReference);
    error PaymentAuthorizationExpired(uint256 deadline);
    error InvalidPaymentAuthorization();
    error InsufficientEthSent(uint256 requiredWei, uint256 sentWei);
    error InvalidEthUsdPrice(int256 price);
    error StaleEthUsdPrice(uint256 updatedAt);
    error InvalidMaximumPriceAge(uint256 maximumPriceAge);
    error UnsupportedTokenDecimals(uint8 decimals);
    error OwnershipCannotBeRenounced();

    constructor(
        address initialOwner,
        address initialTreasury,
        address initialPaymentSigner,
        IERC20Metadata acceptedToken,
        IEthUsdPriceFeed priceFeed,
        uint256 initialMaximumPriceAge
    ) Ownable(initialOwner) EIP712("StreetBornChapterPaymentVault", "1") {
        if (
            initialTreasury == address(0) || initialPaymentSigner == address(0) || address(acceptedToken) == address(0)
                || address(priceFeed) == address(0)
        ) {
            revert ZeroAddress();
        }
        uint8 tokenDecimals = acceptedToken.decimals();
        if (tokenDecimals < CENT_DECIMALS) {
            revert UnsupportedTokenDecimals(tokenDecimals);
        }
        paymentToken = IERC20(address(acceptedToken));
        ethUsdPriceFeed = priceFeed;
        tokenUnitsPerCent = 10 ** (tokenDecimals - CENT_DECIMALS);
        priceFeedScale = 10 ** priceFeed.decimals();
        _setTreasury(initialTreasury);
        _setPaymentSigner(initialPaymentSigner);
        _setMaximumPriceAge(initialMaximumPriceAge);
    }

    function payWithEth(PaymentAuthorization calldata authorization) external payable whenNotPaused nonReentrant {
        _acceptAuthorization(authorization);
        uint256 requiredWei = quoteWei(authorization.usdCents);
        if (msg.value < requiredWei) {
            revert InsufficientEthSent(requiredWei, msg.value);
        }
        emit PaymentReceived(
            authorization.paymentReference, msg.sender, address(0), requiredWei, authorization.usdCents
        );
        uint256 excessWei = msg.value - requiredWei;
        if (excessWei > 0) {
            Address.sendValue(payable(msg.sender), excessWei);
        }
    }

    function payWithToken(PaymentAuthorization calldata authorization) external whenNotPaused nonReentrant {
        _acceptAuthorization(authorization);
        uint256 tokenAmount = quoteTokenUnits(authorization.usdCents);
        paymentToken.safeTransferFrom(msg.sender, address(this), tokenAmount);
        emit PaymentReceived(
            authorization.paymentReference, msg.sender, address(paymentToken), tokenAmount, authorization.usdCents
        );
    }

    function quoteWei(uint256 usdCents) public view returns (uint256) {
        (, int256 price,, uint256 updatedAt,) = ethUsdPriceFeed.latestRoundData();
        if (price <= 0) {
            revert InvalidEthUsdPrice(price);
        }
        if (updatedAt > block.timestamp || block.timestamp - updatedAt > maximumPriceAge) {
            revert StaleEthUsdPrice(updatedAt);
        }
        return Math.mulDiv(
            usdCents, WEI_PER_ETHER * priceFeedScale, SafeCast.toUint256(price) * CENTS_PER_DOLLAR, Math.Rounding.Ceil
        );
    }

    function quoteTokenUnits(uint256 usdCents) public view returns (uint256) {
        return usdCents * tokenUnitsPerCent;
    }

    function hashPaymentAuthorization(bytes32 paymentReference, address payer, uint256 usdCents, uint256 deadline)
        public
        view
        returns (bytes32)
    {
        return
            _hashTypedDataV4(
                keccak256(abi.encode(CHAPTER_PAYMENT_TYPEHASH, paymentReference, payer, usdCents, deadline))
            );
    }

    function setPaymentSigner(address newPaymentSigner) external onlyOwner {
        _setPaymentSigner(newPaymentSigner);
    }

    function setTreasury(address newTreasury) external onlyOwner {
        _setTreasury(newTreasury);
    }

    function setMaximumPriceAge(uint256 newMaximumPriceAge) external onlyOwner {
        _setMaximumPriceAge(newMaximumPriceAge);
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    function withdrawEth(uint256 amount) external onlyOwner nonReentrant {
        emit EthWithdrawn(treasury, amount);
        Address.sendValue(payable(treasury), amount);
    }

    function withdrawToken(IERC20 token, uint256 amount) external onlyOwner nonReentrant {
        emit TokenWithdrawn(address(token), treasury, amount);
        token.safeTransfer(treasury, amount);
    }

    function renounceOwnership() public view override onlyOwner {
        revert OwnershipCannotBeRenounced();
    }

    function _acceptAuthorization(PaymentAuthorization calldata authorization) private {
        if (block.timestamp > authorization.deadline) {
            revert PaymentAuthorizationExpired(authorization.deadline);
        }
        if (authorization.usdCents == 0) {
            revert ZeroPaymentAmount();
        }
        if (isPaymentReferencePaid[authorization.paymentReference]) {
            revert PaymentAlreadyMade(authorization.paymentReference);
        }
        bytes32 digest = hashPaymentAuthorization(
            authorization.paymentReference, msg.sender, authorization.usdCents, authorization.deadline
        );
        (address recoveredSigner, ECDSA.RecoverError recoverError,) = ECDSA.tryRecover(digest, authorization.signature);
        if (recoverError != ECDSA.RecoverError.NoError || recoveredSigner != paymentSigner) {
            revert InvalidPaymentAuthorization();
        }
        isPaymentReferencePaid[authorization.paymentReference] = true;
    }

    function _setPaymentSigner(address newPaymentSigner) private {
        if (newPaymentSigner == address(0)) {
            revert ZeroAddress();
        }
        emit PaymentSignerChanged(paymentSigner, newPaymentSigner);
        paymentSigner = newPaymentSigner;
    }

    function _setTreasury(address newTreasury) private {
        if (newTreasury == address(0)) {
            revert ZeroAddress();
        }
        emit TreasuryChanged(treasury, newTreasury);
        treasury = newTreasury;
    }

    function _setMaximumPriceAge(uint256 newMaximumPriceAge) private {
        if (newMaximumPriceAge == 0 || newMaximumPriceAge > MAXIMUM_PRICE_AGE_LIMIT) {
            revert InvalidMaximumPriceAge(newMaximumPriceAge);
        }
        emit MaximumPriceAgeChanged(maximumPriceAge, newMaximumPriceAge);
        maximumPriceAge = newMaximumPriceAge;
    }
}
