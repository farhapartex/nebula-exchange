// SPDX-License-Identifier: MIT
pragma solidity 0.8.37;

import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import {Ownable2StepUpgradeable} from "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {EIP712Upgradeable} from "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Address} from "@openzeppelin/contracts/utils/Address.sol";
import {ReentrancyGuardTransient} from "@openzeppelin/contracts/utils/ReentrancyGuardTransient.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

contract ChapterPaymentVault is
    Initializable,
    UUPSUpgradeable,
    Ownable2StepUpgradeable,
    PausableUpgradeable,
    EIP712Upgradeable,
    ReentrancyGuardTransient
{
    using SafeERC20 for IERC20;

    struct PaymentAuthorization {
        bytes32 paymentReference;
        uint256 usdCents;
        uint256 deadline;
        bytes signature;
    }

    struct InitializationSettings {
        address owner;
        address treasury;
        address paymentSigner;
        IERC20Metadata paymentToken;
        AggregatorV3Interface ethUsdPriceFeed;
        AggregatorV3Interface sequencerUptimeFeed;
        uint256 maximumPriceAge;
    }

    struct ChapterPaymentVaultStorage {
        IERC20 paymentToken;
        uint256 tokenUnitsPerCent;
        AggregatorV3Interface ethUsdPriceFeed;
        uint256 priceFeedScale;
        AggregatorV3Interface sequencerUptimeFeed;
        address paymentSigner;
        address treasury;
        uint256 maximumPriceAge;
        mapping(bytes32 paymentReference => bool isPaid) isPaymentReferencePaid;
    }

    bytes32 private constant CHAPTER_PAYMENT_VAULT_STORAGE_LOCATION =
        0x42a035f48a8fad60dab5f6f94d505c31e61f1611ed131d1d3fe830cb224d1700;
    bytes32 public constant CHAPTER_PAYMENT_TYPEHASH =
        keccak256("ChapterPayment(bytes32 paymentReference,address payer,uint256 usdCents,uint256 deadline)");
    string private constant SIGNING_DOMAIN_NAME = "StreetBornChapterPaymentVault";
    string private constant SIGNING_DOMAIN_VERSION = "1";
    uint256 public constant MAXIMUM_PRICE_AGE_LIMIT = 1 days;
    uint256 public constant SEQUENCER_GRACE_PERIOD = 1 hours;
    uint256 private constant CENTS_PER_DOLLAR = 100;
    uint256 private constant WEI_PER_ETHER = 1e18;
    uint8 private constant CENT_DECIMALS = 2;
    int256 private constant SEQUENCER_UP = 0;

    event PaymentReceived(
        bytes32 indexed paymentReference,
        address indexed payer,
        address indexed asset,
        uint256 amountPaid,
        uint256 usdCents
    );
    event PaymentSignerChanged(address indexed previousPaymentSigner, address indexed newPaymentSigner);
    event TreasuryChanged(address indexed previousTreasury, address indexed newTreasury);
    event EthUsdPriceFeedChanged(address indexed previousPriceFeed, address indexed newPriceFeed);
    event SequencerUptimeFeedChanged(address indexed previousUptimeFeed, address indexed newUptimeFeed);
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
    error SequencerDown();
    error SequencerGracePeriodNotOver(uint256 sequencerStartedAt);
    error InvalidMaximumPriceAge(uint256 maximumPriceAge);
    error UnsupportedTokenDecimals(uint8 decimals);
    error OwnershipCannotBeRenounced();

    /// @dev Locks the implementation contract so only the proxy can ever be initialized.
    constructor() {
        _disableInitializers();
    }

    /// @notice Sets up the proxy once with the owner, treasury, payment signer, USDC token, price feeds and price age limit.
    function initialize(InitializationSettings calldata settings) external initializer {
        __Ownable_init(settings.owner);
        __Ownable2Step_init();
        __Pausable_init();
        __EIP712_init(SIGNING_DOMAIN_NAME, SIGNING_DOMAIN_VERSION);
        _setPaymentToken(settings.paymentToken);
        _setEthUsdPriceFeed(settings.ethUsdPriceFeed);
        _setSequencerUptimeFeed(settings.sequencerUptimeFeed);
        _setTreasury(settings.treasury);
        _setPaymentSigner(settings.paymentSigner);
        _setMaximumPriceAge(settings.maximumPriceAge);
    }

    /// @notice Pays a signed checkout in ETH at the Chainlink ETH/USD price and refunds any extra ETH sent.
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

    /// @notice Pays a signed checkout with exactly the USD amount in the payment token (USDC).
    function payWithToken(PaymentAuthorization calldata authorization) external whenNotPaused nonReentrant {
        _acceptAuthorization(authorization);
        IERC20 acceptedToken = _vaultStorage().paymentToken;
        uint256 tokenAmount = quoteTokenUnits(authorization.usdCents);
        acceptedToken.safeTransferFrom(msg.sender, address(this), tokenAmount);
        emit PaymentReceived(
            authorization.paymentReference, msg.sender, address(acceptedToken), tokenAmount, authorization.usdCents
        );
    }

    /// @notice Returns how much ETH (in wei) the given USD cents cost at the current Chainlink price, rounded up.
    function quoteWei(uint256 usdCents) public view returns (uint256) {
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        _requireSequencerUp(vaultStorage.sequencerUptimeFeed);
        (, int256 price,, uint256 updatedAt,) = vaultStorage.ethUsdPriceFeed.latestRoundData();
        if (price <= 0) {
            revert InvalidEthUsdPrice(price);
        }
        if (updatedAt == 0 || updatedAt > block.timestamp || block.timestamp - updatedAt > vaultStorage.maximumPriceAge)
        {
            revert StaleEthUsdPrice(updatedAt);
        }
        return Math.mulDiv(
            usdCents,
            WEI_PER_ETHER * vaultStorage.priceFeedScale,
            SafeCast.toUint256(price) * CENTS_PER_DOLLAR,
            Math.Rounding.Ceil
        );
    }

    /// @notice Returns how many payment token units the given USD cents cost.
    function quoteTokenUnits(uint256 usdCents) public view returns (uint256) {
        return usdCents * _vaultStorage().tokenUnitsPerCent;
    }

    /// @notice Returns the EIP-712 digest the payment signer signs to authorize a checkout for a payer.
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

    /// @notice Returns the version number of the logic contract behind the proxy.
    function implementationVersion() external pure virtual returns (uint256) {
        return 1;
    }

    /// @notice Returns the token accepted by payWithToken.
    function paymentToken() external view returns (IERC20) {
        return _vaultStorage().paymentToken;
    }

    /// @notice Returns the Chainlink ETH/USD price feed used to price ETH payments.
    function ethUsdPriceFeed() external view returns (AggregatorV3Interface) {
        return _vaultStorage().ethUsdPriceFeed;
    }

    /// @notice Returns the Chainlink L2 sequencer uptime feed, or the zero address when the check is off.
    function sequencerUptimeFeed() external view returns (AggregatorV3Interface) {
        return _vaultStorage().sequencerUptimeFeed;
    }

    /// @notice Returns the backend address whose signature every payment needs.
    function paymentSigner() external view returns (address) {
        return _vaultStorage().paymentSigner;
    }

    /// @notice Returns the address that receives all withdrawals.
    function treasury() external view returns (address) {
        return _vaultStorage().treasury;
    }

    /// @notice Returns how old, in seconds, the Chainlink price may be before ETH payments are refused.
    function maximumPriceAge() external view returns (uint256) {
        return _vaultStorage().maximumPriceAge;
    }

    /// @notice Returns true when the given checkout reference has already been paid.
    function isPaymentReferencePaid(bytes32 paymentReference) external view returns (bool) {
        return _vaultStorage().isPaymentReferencePaid[paymentReference];
    }

    /// @notice Lets the owner replace the backend payment signer.
    function setPaymentSigner(address newPaymentSigner) external onlyOwner {
        _setPaymentSigner(newPaymentSigner);
    }

    /// @notice Lets the owner change the address that receives withdrawals.
    function setTreasury(address newTreasury) external onlyOwner {
        _setTreasury(newTreasury);
    }

    /// @notice Lets the owner switch to a different Chainlink ETH/USD price feed.
    function setEthUsdPriceFeed(AggregatorV3Interface newPriceFeed) external onlyOwner {
        _setEthUsdPriceFeed(newPriceFeed);
    }

    /// @notice Lets the owner set or remove the Chainlink L2 sequencer uptime feed.
    function setSequencerUptimeFeed(AggregatorV3Interface newUptimeFeed) external onlyOwner {
        _setSequencerUptimeFeed(newUptimeFeed);
    }

    /// @notice Lets the owner change the maximum allowed age of the Chainlink price, up to one day.
    function setMaximumPriceAge(uint256 newMaximumPriceAge) external onlyOwner {
        _setMaximumPriceAge(newMaximumPriceAge);
    }

    /// @notice Lets the owner stop all payments.
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Lets the owner resume payments.
    function unpause() external onlyOwner {
        _unpause();
    }

    /// @notice Lets the owner send ETH from the vault to the treasury.
    function withdrawEth(uint256 amount) external onlyOwner nonReentrant {
        address currentTreasury = _vaultStorage().treasury;
        emit EthWithdrawn(currentTreasury, amount);
        Address.sendValue(payable(currentTreasury), amount);
    }

    /// @notice Lets the owner send any token held by the vault to the treasury.
    function withdrawToken(IERC20 token, uint256 amount) external onlyOwner nonReentrant {
        address currentTreasury = _vaultStorage().treasury;
        emit TokenWithdrawn(address(token), currentTreasury, amount);
        token.safeTransfer(currentTreasury, amount);
    }

    /// @notice Always reverts so the vault can never be left without an owner.
    function renounceOwnership() public view override onlyOwner {
        revert OwnershipCannotBeRenounced();
    }

    /// @dev Allows only the owner to upgrade the proxy to a new implementation.
    function _authorizeUpgrade(address) internal override onlyOwner {}

    /// @dev Checks the deadline, amount, reuse and signer of a payment authorization, then marks the reference paid.
    function _acceptAuthorization(PaymentAuthorization calldata authorization) private {
        if (block.timestamp > authorization.deadline) {
            revert PaymentAuthorizationExpired(authorization.deadline);
        }
        if (authorization.usdCents == 0) {
            revert ZeroPaymentAmount();
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        if (vaultStorage.isPaymentReferencePaid[authorization.paymentReference]) {
            revert PaymentAlreadyMade(authorization.paymentReference);
        }
        bytes32 digest = hashPaymentAuthorization(
            authorization.paymentReference, msg.sender, authorization.usdCents, authorization.deadline
        );
        (address recoveredSigner, ECDSA.RecoverError recoverError,) = ECDSA.tryRecover(digest, authorization.signature);
        if (recoverError != ECDSA.RecoverError.NoError || recoveredSigner != vaultStorage.paymentSigner) {
            revert InvalidPaymentAuthorization();
        }
        vaultStorage.isPaymentReferencePaid[authorization.paymentReference] = true;
    }

    /// @dev Reverts when the L2 sequencer is down or restarted less than the grace period ago.
    function _requireSequencerUp(AggregatorV3Interface uptimeFeed) private view {
        if (address(uptimeFeed) == address(0)) {
            return;
        }
        (, int256 sequencerStatus, uint256 sequencerStartedAt,,) = uptimeFeed.latestRoundData();
        if (sequencerStatus != SEQUENCER_UP) {
            revert SequencerDown();
        }
        if (sequencerStartedAt == 0 || block.timestamp - sequencerStartedAt <= SEQUENCER_GRACE_PERIOD) {
            revert SequencerGracePeriodNotOver(sequencerStartedAt);
        }
    }

    /// @dev Stores the payment token and how many of its units make one cent.
    function _setPaymentToken(IERC20Metadata newPaymentToken) private {
        if (address(newPaymentToken) == address(0)) {
            revert ZeroAddress();
        }
        uint8 tokenDecimals = newPaymentToken.decimals();
        if (tokenDecimals < CENT_DECIMALS) {
            revert UnsupportedTokenDecimals(tokenDecimals);
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        vaultStorage.paymentToken = IERC20(address(newPaymentToken));
        vaultStorage.tokenUnitsPerCent = 10 ** (tokenDecimals - CENT_DECIMALS);
    }

    /// @dev Stores the ETH/USD price feed and its decimal scale.
    function _setEthUsdPriceFeed(AggregatorV3Interface newPriceFeed) private {
        if (address(newPriceFeed) == address(0)) {
            revert ZeroAddress();
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        emit EthUsdPriceFeedChanged(address(vaultStorage.ethUsdPriceFeed), address(newPriceFeed));
        vaultStorage.ethUsdPriceFeed = newPriceFeed;
        vaultStorage.priceFeedScale = 10 ** newPriceFeed.decimals();
    }

    /// @dev Stores the L2 sequencer uptime feed.
    function _setSequencerUptimeFeed(AggregatorV3Interface newUptimeFeed) private {
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        emit SequencerUptimeFeedChanged(address(vaultStorage.sequencerUptimeFeed), address(newUptimeFeed));
        vaultStorage.sequencerUptimeFeed = newUptimeFeed;
    }

    /// @dev Stores a non-zero payment signer address.
    function _setPaymentSigner(address newPaymentSigner) private {
        if (newPaymentSigner == address(0)) {
            revert ZeroAddress();
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        emit PaymentSignerChanged(vaultStorage.paymentSigner, newPaymentSigner);
        vaultStorage.paymentSigner = newPaymentSigner;
    }

    /// @dev Stores a non-zero treasury address.
    function _setTreasury(address newTreasury) private {
        if (newTreasury == address(0)) {
            revert ZeroAddress();
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        emit TreasuryChanged(vaultStorage.treasury, newTreasury);
        vaultStorage.treasury = newTreasury;
    }

    /// @dev Stores a price age limit between one second and one day.
    function _setMaximumPriceAge(uint256 newMaximumPriceAge) private {
        if (newMaximumPriceAge == 0 || newMaximumPriceAge > MAXIMUM_PRICE_AGE_LIMIT) {
            revert InvalidMaximumPriceAge(newMaximumPriceAge);
        }
        ChapterPaymentVaultStorage storage vaultStorage = _vaultStorage();
        emit MaximumPriceAgeChanged(vaultStorage.maximumPriceAge, newMaximumPriceAge);
        vaultStorage.maximumPriceAge = newMaximumPriceAge;
    }

    /// @dev Returns the vault's ERC-7201 namespaced storage struct.
    function _vaultStorage() private pure returns (ChapterPaymentVaultStorage storage vaultStorage) {
        assembly {
            vaultStorage.slot := CHAPTER_PAYMENT_VAULT_STORAGE_LOCATION
        }
    }
}
