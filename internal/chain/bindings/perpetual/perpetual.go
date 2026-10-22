// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package perpetual

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// PerpetualMetaData contains all meta data concerning the Perpetual contract.
var PerpetualMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"paper\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"credit\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getFundingRate\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"liquidate\",\"inputs\":[{\"name\":\"liquidator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"liquidatedTrader\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"requestPaper\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"expectCredit\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"outputs\":[{\"name\":\"liqtorPaperChange\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"liqtorCreditChange\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"trade\",\"inputs\":[{\"name\":\"tradeData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateFundingRate\",\"inputs\":[{\"name\":\"newFundingRate\",\"type\":\"int256\",\"internalType\":\"int256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"BalanceChange\",\"inputs\":[{\"name\":\"trader\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"paperChange\",\"type\":\"int256\",\"indexed\":false,\"internalType\":\"int256\"},{\"name\":\"creditChange\",\"type\":\"int256\",\"indexed\":false,\"internalType\":\"int256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UpdateFundingRate\",\"inputs\":[{\"name\":\"oldFundingRate\",\"type\":\"int256\",\"indexed\":false,\"internalType\":\"int256\"},{\"name\":\"newFundingRate\",\"type\":\"int256\",\"indexed\":false,\"internalType\":\"int256\"}],\"anonymous\":false}]",
}

// PerpetualABI is the input ABI used to generate the binding from.
// Deprecated: Use PerpetualMetaData.ABI instead.
var PerpetualABI = PerpetualMetaData.ABI

// Perpetual is an auto generated Go binding around an Ethereum contract.
type Perpetual struct {
	PerpetualCaller     // Read-only binding to the contract
	PerpetualTransactor // Write-only binding to the contract
	PerpetualFilterer   // Log filterer for contract events
}

// PerpetualCaller is an auto generated read-only Go binding around an Ethereum contract.
type PerpetualCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerpetualTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PerpetualTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerpetualFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PerpetualFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PerpetualSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PerpetualSession struct {
	Contract     *Perpetual        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PerpetualCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PerpetualCallerSession struct {
	Contract *PerpetualCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// PerpetualTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PerpetualTransactorSession struct {
	Contract     *PerpetualTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// PerpetualRaw is an auto generated low-level Go binding around an Ethereum contract.
type PerpetualRaw struct {
	Contract *Perpetual // Generic contract binding to access the raw methods on
}

// PerpetualCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PerpetualCallerRaw struct {
	Contract *PerpetualCaller // Generic read-only contract binding to access the raw methods on
}

// PerpetualTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PerpetualTransactorRaw struct {
	Contract *PerpetualTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPerpetual creates a new instance of Perpetual, bound to a specific deployed contract.
func NewPerpetual(address common.Address, backend bind.ContractBackend) (*Perpetual, error) {
	contract, err := bindPerpetual(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Perpetual{PerpetualCaller: PerpetualCaller{contract: contract}, PerpetualTransactor: PerpetualTransactor{contract: contract}, PerpetualFilterer: PerpetualFilterer{contract: contract}}, nil
}

// NewPerpetualCaller creates a new read-only instance of Perpetual, bound to a specific deployed contract.
func NewPerpetualCaller(address common.Address, caller bind.ContractCaller) (*PerpetualCaller, error) {
	contract, err := bindPerpetual(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PerpetualCaller{contract: contract}, nil
}

// NewPerpetualTransactor creates a new write-only instance of Perpetual, bound to a specific deployed contract.
func NewPerpetualTransactor(address common.Address, transactor bind.ContractTransactor) (*PerpetualTransactor, error) {
	contract, err := bindPerpetual(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PerpetualTransactor{contract: contract}, nil
}

// NewPerpetualFilterer creates a new log filterer instance of Perpetual, bound to a specific deployed contract.
func NewPerpetualFilterer(address common.Address, filterer bind.ContractFilterer) (*PerpetualFilterer, error) {
	contract, err := bindPerpetual(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PerpetualFilterer{contract: contract}, nil
}

// bindPerpetual binds a generic wrapper to an already deployed contract.
func bindPerpetual(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PerpetualMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Perpetual *PerpetualRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Perpetual.Contract.PerpetualCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Perpetual *PerpetualRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Perpetual.Contract.PerpetualTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Perpetual *PerpetualRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Perpetual.Contract.PerpetualTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Perpetual *PerpetualCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Perpetual.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Perpetual *PerpetualTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Perpetual.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Perpetual *PerpetualTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Perpetual.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address trader) view returns(int256 paper, int256 credit)
func (_Perpetual *PerpetualCaller) BalanceOf(opts *bind.CallOpts, trader common.Address) (struct {
	Paper  *big.Int
	Credit *big.Int
}, error) {
	var out []interface{}
	err := _Perpetual.contract.Call(opts, &out, "balanceOf", trader)

	outstruct := new(struct {
		Paper  *big.Int
		Credit *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Paper = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Credit = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address trader) view returns(int256 paper, int256 credit)
func (_Perpetual *PerpetualSession) BalanceOf(trader common.Address) (struct {
	Paper  *big.Int
	Credit *big.Int
}, error) {
	return _Perpetual.Contract.BalanceOf(&_Perpetual.CallOpts, trader)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address trader) view returns(int256 paper, int256 credit)
func (_Perpetual *PerpetualCallerSession) BalanceOf(trader common.Address) (struct {
	Paper  *big.Int
	Credit *big.Int
}, error) {
	return _Perpetual.Contract.BalanceOf(&_Perpetual.CallOpts, trader)
}

// GetFundingRate is a free data retrieval call binding the contract method 0x95196f7e.
//
// Solidity: function getFundingRate() view returns(int256)
func (_Perpetual *PerpetualCaller) GetFundingRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Perpetual.contract.Call(opts, &out, "getFundingRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetFundingRate is a free data retrieval call binding the contract method 0x95196f7e.
//
// Solidity: function getFundingRate() view returns(int256)
func (_Perpetual *PerpetualSession) GetFundingRate() (*big.Int, error) {
	return _Perpetual.Contract.GetFundingRate(&_Perpetual.CallOpts)
}

// GetFundingRate is a free data retrieval call binding the contract method 0x95196f7e.
//
// Solidity: function getFundingRate() view returns(int256)
func (_Perpetual *PerpetualCallerSession) GetFundingRate() (*big.Int, error) {
	return _Perpetual.Contract.GetFundingRate(&_Perpetual.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Perpetual *PerpetualCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Perpetual.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Perpetual *PerpetualSession) Owner() (common.Address, error) {
	return _Perpetual.Contract.Owner(&_Perpetual.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Perpetual *PerpetualCallerSession) Owner() (common.Address, error) {
	return _Perpetual.Contract.Owner(&_Perpetual.CallOpts)
}

// Liquidate is a paid mutator transaction binding the contract method 0xc1790370.
//
// Solidity: function liquidate(address liquidator, address liquidatedTrader, int256 requestPaper, int256 expectCredit) returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Perpetual *PerpetualTransactor) Liquidate(opts *bind.TransactOpts, liquidator common.Address, liquidatedTrader common.Address, requestPaper *big.Int, expectCredit *big.Int) (*types.Transaction, error) {
	return _Perpetual.contract.Transact(opts, "liquidate", liquidator, liquidatedTrader, requestPaper, expectCredit)
}

// Liquidate is a paid mutator transaction binding the contract method 0xc1790370.
//
// Solidity: function liquidate(address liquidator, address liquidatedTrader, int256 requestPaper, int256 expectCredit) returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Perpetual *PerpetualSession) Liquidate(liquidator common.Address, liquidatedTrader common.Address, requestPaper *big.Int, expectCredit *big.Int) (*types.Transaction, error) {
	return _Perpetual.Contract.Liquidate(&_Perpetual.TransactOpts, liquidator, liquidatedTrader, requestPaper, expectCredit)
}

// Liquidate is a paid mutator transaction binding the contract method 0xc1790370.
//
// Solidity: function liquidate(address liquidator, address liquidatedTrader, int256 requestPaper, int256 expectCredit) returns(int256 liqtorPaperChange, int256 liqtorCreditChange)
func (_Perpetual *PerpetualTransactorSession) Liquidate(liquidator common.Address, liquidatedTrader common.Address, requestPaper *big.Int, expectCredit *big.Int) (*types.Transaction, error) {
	return _Perpetual.Contract.Liquidate(&_Perpetual.TransactOpts, liquidator, liquidatedTrader, requestPaper, expectCredit)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Perpetual *PerpetualTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Perpetual.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Perpetual *PerpetualSession) RenounceOwnership() (*types.Transaction, error) {
	return _Perpetual.Contract.RenounceOwnership(&_Perpetual.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Perpetual *PerpetualTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Perpetual.Contract.RenounceOwnership(&_Perpetual.TransactOpts)
}

// Trade is a paid mutator transaction binding the contract method 0x1037ea51.
//
// Solidity: function trade(bytes tradeData) returns()
func (_Perpetual *PerpetualTransactor) Trade(opts *bind.TransactOpts, tradeData []byte) (*types.Transaction, error) {
	return _Perpetual.contract.Transact(opts, "trade", tradeData)
}

// Trade is a paid mutator transaction binding the contract method 0x1037ea51.
//
// Solidity: function trade(bytes tradeData) returns()
func (_Perpetual *PerpetualSession) Trade(tradeData []byte) (*types.Transaction, error) {
	return _Perpetual.Contract.Trade(&_Perpetual.TransactOpts, tradeData)
}

// Trade is a paid mutator transaction binding the contract method 0x1037ea51.
//
// Solidity: function trade(bytes tradeData) returns()
func (_Perpetual *PerpetualTransactorSession) Trade(tradeData []byte) (*types.Transaction, error) {
	return _Perpetual.Contract.Trade(&_Perpetual.TransactOpts, tradeData)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Perpetual *PerpetualTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Perpetual.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Perpetual *PerpetualSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Perpetual.Contract.TransferOwnership(&_Perpetual.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Perpetual *PerpetualTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Perpetual.Contract.TransferOwnership(&_Perpetual.TransactOpts, newOwner)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0xc2684de6.
//
// Solidity: function updateFundingRate(int256 newFundingRate) returns()
func (_Perpetual *PerpetualTransactor) UpdateFundingRate(opts *bind.TransactOpts, newFundingRate *big.Int) (*types.Transaction, error) {
	return _Perpetual.contract.Transact(opts, "updateFundingRate", newFundingRate)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0xc2684de6.
//
// Solidity: function updateFundingRate(int256 newFundingRate) returns()
func (_Perpetual *PerpetualSession) UpdateFundingRate(newFundingRate *big.Int) (*types.Transaction, error) {
	return _Perpetual.Contract.UpdateFundingRate(&_Perpetual.TransactOpts, newFundingRate)
}

// UpdateFundingRate is a paid mutator transaction binding the contract method 0xc2684de6.
//
// Solidity: function updateFundingRate(int256 newFundingRate) returns()
func (_Perpetual *PerpetualTransactorSession) UpdateFundingRate(newFundingRate *big.Int) (*types.Transaction, error) {
	return _Perpetual.Contract.UpdateFundingRate(&_Perpetual.TransactOpts, newFundingRate)
}

// PerpetualBalanceChangeIterator is returned from FilterBalanceChange and is used to iterate over the raw logs and unpacked data for BalanceChange events raised by the Perpetual contract.
type PerpetualBalanceChangeIterator struct {
	Event *PerpetualBalanceChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PerpetualBalanceChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PerpetualBalanceChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PerpetualBalanceChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PerpetualBalanceChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PerpetualBalanceChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PerpetualBalanceChange represents a BalanceChange event raised by the Perpetual contract.
type PerpetualBalanceChange struct {
	Trader       common.Address
	PaperChange  *big.Int
	CreditChange *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBalanceChange is a free log retrieval operation binding the contract event 0x4671546415242c95e8c6ef84d4785a1477e86e46fab15b87e504a3da994eba7c.
//
// Solidity: event BalanceChange(address indexed trader, int256 paperChange, int256 creditChange)
func (_Perpetual *PerpetualFilterer) FilterBalanceChange(opts *bind.FilterOpts, trader []common.Address) (*PerpetualBalanceChangeIterator, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Perpetual.contract.FilterLogs(opts, "BalanceChange", traderRule)
	if err != nil {
		return nil, err
	}
	return &PerpetualBalanceChangeIterator{contract: _Perpetual.contract, event: "BalanceChange", logs: logs, sub: sub}, nil
}

// WatchBalanceChange is a free log subscription operation binding the contract event 0x4671546415242c95e8c6ef84d4785a1477e86e46fab15b87e504a3da994eba7c.
//
// Solidity: event BalanceChange(address indexed trader, int256 paperChange, int256 creditChange)
func (_Perpetual *PerpetualFilterer) WatchBalanceChange(opts *bind.WatchOpts, sink chan<- *PerpetualBalanceChange, trader []common.Address) (event.Subscription, error) {

	var traderRule []interface{}
	for _, traderItem := range trader {
		traderRule = append(traderRule, traderItem)
	}

	logs, sub, err := _Perpetual.contract.WatchLogs(opts, "BalanceChange", traderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PerpetualBalanceChange)
				if err := _Perpetual.contract.UnpackLog(event, "BalanceChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBalanceChange is a log parse operation binding the contract event 0x4671546415242c95e8c6ef84d4785a1477e86e46fab15b87e504a3da994eba7c.
//
// Solidity: event BalanceChange(address indexed trader, int256 paperChange, int256 creditChange)
func (_Perpetual *PerpetualFilterer) ParseBalanceChange(log types.Log) (*PerpetualBalanceChange, error) {
	event := new(PerpetualBalanceChange)
	if err := _Perpetual.contract.UnpackLog(event, "BalanceChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PerpetualOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Perpetual contract.
type PerpetualOwnershipTransferredIterator struct {
	Event *PerpetualOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PerpetualOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PerpetualOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PerpetualOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PerpetualOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PerpetualOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PerpetualOwnershipTransferred represents a OwnershipTransferred event raised by the Perpetual contract.
type PerpetualOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Perpetual *PerpetualFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PerpetualOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Perpetual.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PerpetualOwnershipTransferredIterator{contract: _Perpetual.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Perpetual *PerpetualFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PerpetualOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Perpetual.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PerpetualOwnershipTransferred)
				if err := _Perpetual.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Perpetual *PerpetualFilterer) ParseOwnershipTransferred(log types.Log) (*PerpetualOwnershipTransferred, error) {
	event := new(PerpetualOwnershipTransferred)
	if err := _Perpetual.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PerpetualUpdateFundingRateIterator is returned from FilterUpdateFundingRate and is used to iterate over the raw logs and unpacked data for UpdateFundingRate events raised by the Perpetual contract.
type PerpetualUpdateFundingRateIterator struct {
	Event *PerpetualUpdateFundingRate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PerpetualUpdateFundingRateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PerpetualUpdateFundingRate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PerpetualUpdateFundingRate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PerpetualUpdateFundingRateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PerpetualUpdateFundingRateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PerpetualUpdateFundingRate represents a UpdateFundingRate event raised by the Perpetual contract.
type PerpetualUpdateFundingRate struct {
	OldFundingRate *big.Int
	NewFundingRate *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpdateFundingRate is a free log retrieval operation binding the contract event 0x41a4e7eeb93ced398be7dabdc68a1aabb9a4755a11e80587da132e9c09f6e3c5.
//
// Solidity: event UpdateFundingRate(int256 oldFundingRate, int256 newFundingRate)
func (_Perpetual *PerpetualFilterer) FilterUpdateFundingRate(opts *bind.FilterOpts) (*PerpetualUpdateFundingRateIterator, error) {

	logs, sub, err := _Perpetual.contract.FilterLogs(opts, "UpdateFundingRate")
	if err != nil {
		return nil, err
	}
	return &PerpetualUpdateFundingRateIterator{contract: _Perpetual.contract, event: "UpdateFundingRate", logs: logs, sub: sub}, nil
}

// WatchUpdateFundingRate is a free log subscription operation binding the contract event 0x41a4e7eeb93ced398be7dabdc68a1aabb9a4755a11e80587da132e9c09f6e3c5.
//
// Solidity: event UpdateFundingRate(int256 oldFundingRate, int256 newFundingRate)
func (_Perpetual *PerpetualFilterer) WatchUpdateFundingRate(opts *bind.WatchOpts, sink chan<- *PerpetualUpdateFundingRate) (event.Subscription, error) {

	logs, sub, err := _Perpetual.contract.WatchLogs(opts, "UpdateFundingRate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PerpetualUpdateFundingRate)
				if err := _Perpetual.contract.UnpackLog(event, "UpdateFundingRate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdateFundingRate is a log parse operation binding the contract event 0x41a4e7eeb93ced398be7dabdc68a1aabb9a4755a11e80587da132e9c09f6e3c5.
//
// Solidity: event UpdateFundingRate(int256 oldFundingRate, int256 newFundingRate)
func (_Perpetual *PerpetualFilterer) ParseUpdateFundingRate(log types.Log) (*PerpetualUpdateFundingRate, error) {
	event := new(PerpetualUpdateFundingRate)
	if err := _Perpetual.contract.UnpackLog(event, "UpdateFundingRate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
