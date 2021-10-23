//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.4;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

import "hardhat/console.sol";
import "./interfaces/IPaymentGateway.sol";
import "./FilecoinOracle.sol";

contract SwanPayment is IPaymentMinimal, Initializable {

    address public constant NATIVE_TOKEN = 0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE;

    address private _owner;
    address private _oracle;
    address private _priceFeed;
    uint256 lockTime = 1 days;
    mapping(string => TxInfo) txMap;

    
    function initialize(address owner) public initializer {
        _owner = owner;
    }


    constructor(address owner) public {
        _owner = owner;
    }

    /**
     * @dev Throws if called by any account other than the owner.
     */
    modifier onlyOwner() {
        require(_owner == msg.sender, "Caller is not the owner");
        _;
    }

    function setOracle(address oracle) public onlyOwner returns (bool) {
        _oracle = oracle;
        return true;
    }

    // /**
    //  * @dev Throws if called by any account other than the owner.
    //  */
    // modifier onlyParticipant() {
    //     require(_owner == msg.sender, "Caller is not the owner");
    //     _;
    // }

    function getLockedPaymentInfo(string calldata txId)
        public
        view
        override
        returns (TxInfo memory tx)
    {
        // default value is 0
        return txMap[txId];
    }


    event LockPayment(
        string id,
        address token,
        uint256 lockedFee,
        uint256 minPayment,
        address recipient,
        uint256 deadline
    );

    /// @notice Deposits the amount of token for specific transaction
    /// @param param The transaction information for which to deposit balance
    /// @return Returns true for a successful deposit, false for an unsuccessful deposit
    function lockPayment(lockPaymentParam calldata param)
        public
        payable
        override
        returns (bool)
    {
        require(!txMap[param.id]._isExisted, "Payment of transaction is already locked");
        require(param.minPayment > 0 && msg.value > param.minPayment, "payment should greater than min payment");
        TxInfo storage t = txMap[param.id];
        t.owner = msg.sender;
        t.minPayment = param.minPayment;
        t.recipient = param.recipient;
        t.deadline = block.timestamp + param.lockTime;
        t.lockedFee = msg.value;
        t._isExisted = true;

        emit LockPayment(
            param.id,
            NATIVE_TOKEN,
            t.lockedFee,
            param.minPayment,
            param.recipient,
            t.deadline
        );
        return true;
    }

    /// @notice Deposits the amount of token for specific transaction
    /// @param param The transaction information for which to deposit balance
    /// @return Returns true for a successful deposit, false for an unsuccessful deposit
    function lockTokenPayment(lockPaymentParam calldata param)
        public
        returns (bool)
    {
        require(!txMap[param.id]._isExisted, "Payment of transaction is already locked");
        require(param.minPayment > 0 && param.amount > param.minPayment, "payment should greater than min payment");
        TxInfo storage t = txMap[param.id];
        t.owner = msg.sender;
        t.token = param.token;
        t.minPayment = param.minPayment;
        t.recipient = param.recipient;
        t.deadline = block.timestamp + param.lockTime;
        t.lockedFee = param.amount;
        t._isExisted = true;

        emit LockPayment(
            param.id,
            t.token,
            t.lockedFee,
            param.minPayment,
            param.recipient,
            t.deadline
        );
        return true;
    }

    /// @notice Returns the current allowance given to a spender by an owner
    /// @param txId transaction id
    /// @return Returns true for a successful payment, false for an unsuccessful payment
    function unlockPayment(string calldata txId)
        public
        override
        returns (bool)
    {
        TxInfo storage t = txMap[txId];
        require(t._isExisted, "Transaction does not exist");
        require(
            t.owner == msg.sender || t.recipient == msg.sender,
            "Invalid caller"
        );
        // if passed deadline, return payback to user
        if (block.timestamp > t.deadline) {
            require(t.owner == msg.sender, "Tx passed deadline, only owner can get locked tokens");
            t._isExisted = false;

            payable(address(t.owner)).transfer(t.lockedFee);
        } else {
            uint256 actualFee = FilecoinOracle(_oracle).getPaymentInfo(txId);
            require(actualFee > 0, "Transaction is incompleted");


            // todo: add convert rate function to get latest price

            if (actualFee < t.minPayment) {
                actualFee = t.minPayment;
            }
            t._isExisted = false;

            console.log("actualFee is %s", actualFee);
             console.log("locked fee is %s", t.lockedFee);
            
            if (t.lockedFee > actualFee) {
                payable(address(t.owner)).transfer(t.lockedFee - actualFee);
            }else{
                actualFee = t.lockedFee;
            }
            payable(address(t.recipient)).transfer(actualFee);
        }
        t.minPayment = 0;
        t.lockedFee = 0;
        // todo: get status from oralce/other contract, status include status, real fee
        // check status, if not complete, return

        return true;
        // real fee is greater than tx.fee, take tx.fee
        // real fee is less than tx.minPayment, take minPayment, return tx.fee - minPayment to tx.owner
        // otherwise, take real fee, return tx.fee - real fee to tx.owner
    }

    /// @notice Returns the current allowance given to a spender by an owner
    /// @param txId transaction id
    /// @return Returns true for a successful payment, false for an unsuccessful payment
    function unlockTokenPayment(string calldata txId)
        public
        override
        returns (bool)
    {
        TxInfo storage t = txMap[txId];
        require(t._isExisted, "Transaction does not exist");
        require(
            t.owner == msg.sender || t.recipient == msg.sender,
            "Invalid caller"
        );
        // if passed deadline, return payback to user
        if (block.timestamp > t.deadline) {
            require(t.owner == msg.sender, "Tx passed deadline, only owner can get locked tokens");
            t._isExisted = false;

            //payable(address(t.owner)).transfer(t.lockedFee);
            // todo: transfer token back

        } else {
            uint256 actualFee = FilecoinOracle(_oracle).getPaymentInfo(txId);
            require(actualFee > 0, "Transaction is incompleted");


            // todo: add convert rate function to get latest price

            if (actualFee < t.minPayment) {
                actualFee = t.minPayment;
            }
            t._isExisted = false;

            console.log("actualFee is %s", actualFee);
             console.log("locked fee is %s", t.lockedFee);
            
            if (t.lockedFee > actualFee) {
                // payable(address(t.owner)).transfer(t.lockedFee - actualFee);
            }else{
                actualFee = t.lockedFee;
            }
            // payable(address(t.recipient)).transfer(actualFee);
        }
        t.minPayment = 0;
        t.lockedFee = 0;
        // todo: get status from oralce/other contract, status include status, real fee
        // check status, if not complete, return

        return true;
        // real fee is greater than tx.fee, take tx.fee
        // real fee is less than tx.minPayment, take minPayment, return tx.fee - minPayment to tx.owner
        // otherwise, take real fee, return tx.fee - real fee to tx.owner
    }

    // function greet() public view returns (string memory) {
    //   return greeting;
    // }

    // function setGreeting(string memory _greeting) public {
    //   console.log("Changing greeting from '%s' to '%s'", greeting, _greeting);
    //   greeting = _greeting;
    // }
}
