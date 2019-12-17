package contracts

import (
	"bytes"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/contracts/common"
	"github.com/vitelabs/go-vite/vm/contracts/defi"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_db"
	"math/big"
)

var deFiLogger = log15.New("module", "deFi")

type MethodDeFiDeposit struct {
	MethodName string
}

func (md *MethodDeFiDeposit) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiDeposit) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiDeposit) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiDeposit) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiDepositQuota
}

func (md *MethodDeFiDeposit) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.Amount.Sign() <= 0 || block.TokenId != ledger.ViteTokenId {
		return defi.InvalidInputParamErr
	}
	return nil
}

func (md *MethodDeFiDeposit) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) ([]*ledger.AccountBlock, error) {
	defi.OnAccDeposit(db, sendBlock.AccountAddress, sendBlock.TokenId, sendBlock.Amount)
	return nil, nil
}

type MethodDeFiWithdraw struct {
	MethodName string
}

func (md *MethodDeFiWithdraw) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiWithdraw) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiWithdraw) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiWithdraw) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiWithdrawQuota
}

func (md *MethodDeFiWithdraw) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	var err error
	param := new(defi.ParamWithdraw)
	if err = abi.ABIDeFi.UnpackMethod(param, md.MethodName, block.Data); err != nil {
		return err
	}
	if param.Amount.Sign() <= 0 || block.TokenId != ledger.ViteTokenId {
		return defi.InvalidInputParamErr
	}
	return nil
}

func (md *MethodDeFiWithdraw) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) ([]*ledger.AccountBlock, error) {
	var (
		param = new(defi.ParamWithdraw)
		err   error
	)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	if _, err = defi.OnAccWithdraw(db, sendBlock.AccountAddress, param.Token.Bytes(), param.Amount); err != nil {
		return nil, err
	} else {
		return []*ledger.AccountBlock{
			{
				AccountAddress: types.AddressDeFi,
				ToAddress:      sendBlock.AccountAddress,
				BlockType:      ledger.BlockTypeSendCall,
				Amount:         param.Amount,
				TokenId:        param.Token,
				Data:           []byte{},
			},
		}, nil
	}
}

type MethodDeFiNewLoan struct {
	MethodName string
}

func (md *MethodDeFiNewLoan) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiNewLoan) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiNewLoan) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiNewLoan) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiNewLoanQuota
}

func (md *MethodDeFiNewLoan) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	var err error
	param := new(defi.ParamNewLoan)
	if err = abi.ABIDeFi.UnpackMethod(param, md.MethodName, block.Data); err != nil {
		return err
	}
	return defi.CheckLoanParam(param)
}

func (md *MethodDeFiNewLoan) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) ([]*ledger.AccountBlock, error) {
	var (
		param = new(defi.ParamNewLoan)
		err   error
	)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	interest := defi.CalculateInterest(param.Shares, param.ShareAmount, param.DayRate, param.ExpireDays)
	if _, err = defi.OnAccNewLoan(db, sendBlock.AccountAddress, interest); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	} else {
		loan := defi.NewLoan(sendBlock.AccountAddress, db, param, interest)
		defi.SaveLoan(db, loan)
		defi.AddNewLoanEvent(db, loan)
	}
	return nil, nil
}

type MethodDeFiCancelLoan struct {
	MethodName string
}

func (md *MethodDeFiCancelLoan) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiCancelLoan) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiCancelLoan) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiCancelLoan) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiCancelLoanQuota
}

func (md *MethodDeFiCancelLoan) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	return abi.ABIDeFi.UnpackMethod(new(uint64), md.MethodName, block.Data)
}

func (md *MethodDeFiCancelLoan) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) ([]*ledger.AccountBlock, error) {
	loanId := new(uint64)
	abi.ABIDeFi.UnpackMethod(loanId, md.MethodName, sendBlock.Data)
	if loan, ok := defi.GetLoanById(db, *loanId); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.LoanNotExistsErr, sendBlock)
	} else {
		if !bytes.Equal(loan.Address, sendBlock.AccountAddress.Bytes()) {
			return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.OnlyOwnerAllowErr, sendBlock)
		} else if loan.Status != defi.Open || loan.SubscribedShares != 0 {
			return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvalidLoanStatusForCancelErr, sendBlock)
		} else {
			defi.OnAccLoanCancelled(db, sendBlock.AccountAddress, loan.Interest)
			loan.Status = defi.Failed
			loan.Updated = defi.GetDeFiTimestamp(db)
			defi.DeleteLoan(db, loan)
			defi.AddLoanUpdateEvent(db, loan)
		}
	}
	return nil, nil
}

type MethodDeFiSubscribe struct {
	MethodName string
}

func (md *MethodDeFiSubscribe) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiSubscribe) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiSubscribe) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiSubscribe) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiSubscribeQuota
}

func (md *MethodDeFiSubscribe) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	var err error
	param := new(defi.ParamSubscribe)
	if err = abi.ABIDeFi.UnpackMethod(param, md.MethodName, block.Data); err != nil {
		return err
	} else {
		if param.LoanId == 0 || param.Shares <= 0 {
			return defi.InvalidInputParamErr
		}
	}
	return nil
}

func (md *MethodDeFiSubscribe) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) ([]*ledger.AccountBlock, error) {
	var (
		param = new(defi.ParamSubscribe)
		loan  *defi.Loan
		sub   *defi.Subscription
		ok    bool
		err   error
	)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	if loan, ok = defi.GetLoanById(db, param.LoanId); !ok || loan.Status != defi.Open {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.LoanNotExistsErr, sendBlock)
	} else if defi.IsOpenLoanSubscribeFail(db, loan) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.LoanSubscribeFailed, sendBlock)
	}
	leavedShares := loan.Shares - loan.SubscribedShares
	if param.Shares > leavedShares {
		param.Shares = leavedShares
	}
	amount := defi.CalculateAmount(param.Shares, loan.ShareAmount)
	if _, err = defi.OnAccSubscribe(db, sendBlock.AccountAddress, amount); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	if sub, ok = defi.GetSubscription(db, param.LoanId, sendBlock.AccountAddress.Bytes()); !ok {
		sub = defi.NewSubscription(sendBlock.AccountAddress, db, param, loan)
	} else {
		sub.Shares = sub.Shares + param.Shares
		sub.Updated = defi.GetDeFiTimestamp(db)
	}
	defi.SaveSubscription(db, sub)
	defi.AddNewSubscriptionEvent(db, sub)
	defi.DoSubscribe(db, vm.GlobalStatus(), loan, param.Shares)
	return nil, nil
}

type MethodDeFiInvest struct {
	MethodName string
}

func (md *MethodDeFiInvest) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiInvest) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiInvest) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiInvest) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiInvestQuota
}

func (md *MethodDeFiInvest) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	var err error
	param := new(defi.ParamInvest)
	if err = abi.ABIDeFi.UnpackMethod(param, md.MethodName, block.Data); err != nil {
		return err
	} else {
		if param.LoanId == 0 || param.BizType < defi.InvestForMining || param.BizType > defi.InvestForQuota || param.BizType == defi.InvestForSBP {
			return defi.InvalidInputParamErr
		}
	}
	return nil
}

func (md *MethodDeFiInvest) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var (
		param                      = new(defi.ParamInvest)
		loanInvested, baseInvested *big.Int
		durationHeight             uint64
		loan                       *defi.Loan
		ok                         bool
	)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	if loan, ok = defi.GetLoanById(db, param.LoanId); !ok || loan.Status != defi.Success {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.LoanNotExistsErr, sendBlock)
	} else if !bytes.Equal(loan.Address, sendBlock.AccountAddress.Bytes()) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.OnlyOwnerAllowErr, sendBlock)
	}
	leavedAmount, availableHeight := defi.GetLoanAvailable(vm.GlobalStatus(), loan)
	if loanInvested, baseInvested, durationHeight, err = defi.PrepareInvest(db, sendBlock.AccountAddress, param.BizType, leavedAmount, param.Amount, availableHeight, nodeConfig.params.StakeHeight, nodeConfig.params.DexSuperVipStakeHeight); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	if _, err = defi.OnAccInvest(db, sendBlock.AccountAddress, loanInvested, baseInvested); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	defi.OnLoanInvest(db, loan, loanInvested)
	invest := defi.NewInvest(db, vm.GlobalStatus(), sendBlock.AccountAddress, loan, param.BizType, param.Beneficiary, loanInvested, baseInvested, durationHeight)
	defi.SaveInvest(db, invest)
	defi.SaveInvestToLoanIndex(db, invest)
	switch param.BizType {
	case defi.InvestForMining:
		blocks, err = defi.DoDexInvest(invest, dex.StakeForMining, sendBlock.Amount)
	case defi.InvestForSVIP:
		blocks, err = defi.DoDexInvest(invest, dex.StakeForPrincipalSuperVIP, sendBlock.Amount)
	case defi.InvestForQuota:
		blocks, err = defi.DoQuotaInvest(db, invest, sendBlock.Amount, nodeConfig.params.StakeHeight, block)
	}
	if err != nil {
		handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	return
}

type MethodDeFiCancelInvest struct {
	MethodName string
}

func (md *MethodDeFiCancelInvest) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiCancelInvest) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiCancelInvest) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiCancelInvest) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiCancelInvestQuota
}

func (md *MethodDeFiCancelInvest) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	var investId = new(uint64)
	if err := abi.ABIDeFi.UnpackMethod(investId, md.MethodName, block.Data); err != nil {
		return err
	} else if *investId == 0 {
		return defi.InvalidInputParamErr
	}
	return nil
}

func (md *MethodDeFiCancelInvest) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var (
		investId  = new(uint64)
		invest    *defi.Invest
		ok        bool
	)
	abi.ABIDeFi.UnpackMethod(investId, md.MethodName, sendBlock.Data)
	if invest, ok = defi.GetInvest(db, *investId); !ok || invest.Status != defi.InvestSuccess {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExistsErr, sendBlock)
	} else if !bytes.Equal(sendBlock.AccountAddress.Bytes(), invest.Address) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.OnlyOwnerAllowErr, sendBlock)
	} else if !defi.IsInvestExpired(vm.GlobalStatus(), invest) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExpiredErr, sendBlock)
	}
	defi.CancellingInvest(db, invest)
	switch invest.BizType {
	case defi.InvestForMining, defi.InvestForSVIP:
		blocks, err = defi.DoCancelDexInvest(invest.Id)
	case defi.InvestForQuota:
		blocks, err = defi.DoCancelQuotaInvest(invest.InvestHash)
	case defi.InvestForSBP:
		blocks, err = defi.DoRevokeSBP(db, invest.InvestHash)
	}
	if err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	return
}

//dex invest InvestForMining, InvestForSVIP
type MethodDeFiRefundInvest struct {
	MethodName string
}

func (md *MethodDeFiRefundInvest) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiRefundInvest) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiRefundInvest) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiRefundInvest) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiRefundInvestQuota
}

func (md *MethodDeFiRefundInvest) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.AccountAddress != types.AddressDexFund {
		return defi.InvalidSourceAddressErr
	}
	var param = new(defi.ParamRefundInvest)
	if err := abi.ABIDeFi.UnpackMethod(param, md.MethodName, block.Data); err != nil {
		return err
	} else if len(param.InvestHashes) == 0 {
		return defi.InvalidInputParamErr
	}
	return nil
}

//handle InvestForMining, InvestForSVIP
func (md *MethodDeFiRefundInvest) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var (
		param = new(defi.ParamRefundInvest)
	)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	for i := 0; i < len(param.InvestHashes)/8; i++ {
		iv := (param.InvestHashes)[i*8 : (i+1)*8]
		investId := common.BytesToUint64(iv)
		if invest, ok := defi.GetInvest(db, investId); !ok {
			panic(defi.InvestNotExistsErr)
		} else {
			defi.DoRefundInvest(db, invest)
		}
	}
	return
}

type MethodDeFiDelegateStakeCallback struct {
	MethodName string
}

func (md *MethodDeFiDelegateStakeCallback) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiDelegateStakeCallback) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiDelegateStakeCallback) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiDelegateStakeCallback) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiDelegateStakeCallbackQuota
}

func (md *MethodDeFiDelegateStakeCallback) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.AccountAddress != types.AddressQuota {
		return defi.InvalidSourceAddressErr
	}
	return abi.ABIDeFi.UnpackMethod(new(dex.ParamDelegateStakeCallbackV2), md.MethodName, block.Data)
}

func (md MethodDeFiDelegateStakeCallback) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var param = new(dex.ParamDelegateStakeCallbackV2)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	var (
		info   *defi.InvestQuotaInfo
		invest *defi.Invest
		ok     bool
	)
	if info, ok = defi.GetInvestQuotaInfo(db, param.Id.Bytes()); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvalidQuotaInvestErr, sendBlock)
	} else if invest, ok = defi.GetInvest(db, info.InvestId); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExistsErr, sendBlock)
	}
	if param.Success {
		defi.ConfirmInvest(db, invest)
	} else {
		defi.DoRefundQuotaInvest(db, invest)
	}
	return
}

type MethodDeFiCancelDelegateStakeCallback struct {
	MethodName string
}

func (md *MethodDeFiCancelDelegateStakeCallback) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiCancelDelegateStakeCallback) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiCancelDelegateStakeCallback) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiCancelDelegateStakeCallback) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiCancelDelegateStakeCallbackQuota
}

func (md *MethodDeFiCancelDelegateStakeCallback) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.AccountAddress != types.AddressQuota {
		return defi.InvalidSourceAddressErr
	}
	return abi.ABIDeFi.UnpackMethod(new(dex.ParamDelegateStakeCallbackV2), md.MethodName, block.Data)
}

func (md MethodDeFiCancelDelegateStakeCallback) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var param = new(dex.ParamDelegateStakeCallbackV2)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	var (
		info   *defi.InvestQuotaInfo
		invest *defi.Invest
		ok     bool
	)
	if info, ok = defi.GetInvestQuotaInfo(db, param.Id.Bytes()); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvalidQuotaInvestErr, sendBlock)
	} else if invest, ok = defi.GetInvest(db, info.InvestId); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExistsErr, sendBlock)
	}
	if param.Success {
		defi.DoRefundQuotaInvest(db, invest)
	}
	return nil, nil
}

type MethodDeFiRegisterSBP struct {
	MethodName string
}

func (md *MethodDeFiRegisterSBP) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiRegisterSBP) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiRegisterSBP) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiRegisterSBP) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiRegisterSBPQuota
}

func (md *MethodDeFiRegisterSBP) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	return abi.ABIDeFi.UnpackMethod(new(defi.ParamRegisterSBP), md.MethodName, block.Data)
}

func (md MethodDeFiRegisterSBP) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var param = new(defi.ParamRegisterSBP)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	var (
		loan                       *defi.Loan
		invest                     *defi.Invest
		loanInvested, baseInvested *big.Int
		durationHeight             uint64
		ok                         bool
	)
	if loan, ok = defi.GetLoanById(db, param.LoanId); !ok || loan.Status != defi.Success {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.LoanNotExistsErr, sendBlock)
	} else if !bytes.Equal(loan.Address, sendBlock.AccountAddress.Bytes()) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.OnlyOwnerAllowErr, sendBlock)
	}
	groupInfo, err := abi.GetConsensusGroup(db, types.SNAPSHOT_GID)
	stakeParam, _ := abi.GetRegisterStakeParamOfConsensusGroup(groupInfo.RegisterConditionParam)
	available, availableHeight := defi.GetLoanAvailable(vm.GlobalStatus(), loan)
	if loanInvested, baseInvested, durationHeight, err = defi.PrepareInvest(db, sendBlock.AccountAddress, defi.InvestForSBP, available, SbpStakeAmountMainnet, availableHeight, stakeParam.StakeHeight, nodeConfig.params.DexSuperVipStakeHeight); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	if _, err = defi.OnAccInvest(db, sendBlock.AccountAddress, loanInvested, baseInvested); err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	}
	defi.OnLoanInvest(db, loan, loanInvested)
	invest = defi.NewInvest(db, vm.GlobalStatus(), sendBlock.AccountAddress, loan, defi.InvestForSBP, types.ZERO_ADDRESS, loanInvested, baseInvested, durationHeight)
	defi.SaveInvestToLoanIndex(db, invest)
	return defi.DoRegisterSBP(db, invest, param, SbpStakeAmountMainnet, block)
}

type MethodDeFiUpdateSBPRegistration struct {
	MethodName string
}

func (md *MethodDeFiUpdateSBPRegistration) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiUpdateSBPRegistration) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiUpdateSBPRegistration) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiUpdateSBPRegistration) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiUpdateSBPRegistrationQuota
}

func (md *MethodDeFiUpdateSBPRegistration) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	return abi.ABIDeFi.UnpackMethod(new(defi.ParamUpdateSBPRegistration), md.MethodName, block.Data)
}

func (md MethodDeFiUpdateSBPRegistration) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var param = new(defi.ParamUpdateSBPRegistration)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	var (
		invest *defi.Invest
		ok     bool
	)
	if invest, ok = defi.GetInvest(db, param.InvestId); !ok || invest.Status != defi.InvestSuccess {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExistsErr, sendBlock)
	} else if !bytes.Equal(sendBlock.AccountAddress.Bytes(), invest.Address) {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.OnlyOwnerAllowErr, sendBlock)
	}
	defi.DoUpdateSBP(db, invest.InvestHash, param)
	return nil, nil
}

type MethodDeFiDefault struct {
	MethodName string
}

func (md *MethodDeFiDefault) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiDefault) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiDefault) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiDefault) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiDefaultQuota
}

func (md *MethodDeFiDefault) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.AccountAddress != types.AddressDexFund && block.AccountAddress != types.AddressGovernance {
		return defi.InvalidSourceAddressErr
	}
	return nil
}

func (md MethodDeFiDefault) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	if originSendBlock := util.GetOriginalSendBlock(db, sendBlock.Hash); originSendBlock != nil && originSendBlock.AccountAddress == types.AddressDeFi && originSendBlock.Amount.Sign() > 0 {
		switch originSendBlock.ToAddress {
		//DelegateInvest
		case types.AddressDexFund:
			blocks, err = defi.HandleDexRefundForException(db, originSendBlock)
			//RegisterSBP, RevokeSBP
		case types.AddressGovernance:
			blocks, err = defi.HandleGovernanceFeedback(db, originSendBlock)
		}
	} else {
		err = defi.InvalidInputParamErr
	}
	if err != nil {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, err, sendBlock)
	} else {
		return
	}
}

type MethodDeFiTriggerJob struct {
	MethodName string
}

func (md *MethodDeFiTriggerJob) GetFee(block *ledger.AccountBlock) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (md *MethodDeFiTriggerJob) GetRefundData(sendBlock *ledger.AccountBlock, sbHeight uint64) ([]byte, bool) {
	return []byte{}, false
}

func (md *MethodDeFiTriggerJob) GetSendQuota(data []byte, gasTable *util.QuotaTable) (uint64, error) {
	return util.RequestQuotaCost(data, gasTable)
}

func (md *MethodDeFiTriggerJob) GetReceiveQuota(gasTable *util.QuotaTable) uint64 {
	return gasTable.DeFiDefaultQuota
}

func (md *MethodDeFiTriggerJob) DoSend(db vm_db.VmDb, block *ledger.AccountBlock) error {
	if block.AccountAddress != types.AddressQuota {
		return defi.InvalidSourceAddressErr
	}
	return abi.ABIDeFi.UnpackMethod(new(dex.ParamDelegateStakeCallbackV2), md.MethodName, block.Data)
}

func (md MethodDeFiTriggerJob) DoReceive(db vm_db.VmDb, block *ledger.AccountBlock, sendBlock *ledger.AccountBlock, vm vmEnvironment) (blocks []*ledger.AccountBlock, err error) {
	var param = new(dex.ParamDelegateStakeCallbackV2)
	abi.ABIDeFi.UnpackMethod(param, md.MethodName, sendBlock.Data)
	var (
		info   *defi.InvestQuotaInfo
		invest *defi.Invest
		ok     bool
	)
	if info, ok = defi.GetInvestQuotaInfo(db, param.Id.Bytes()); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvalidQuotaInvestErr, sendBlock)
	} else if invest, ok = defi.GetInvest(db, info.InvestId); !ok {
		return handleDeFiReceiveErr(deFiLogger, md.MethodName, defi.InvestNotExistsErr, sendBlock)
	}
	if param.Success {
		defi.DoRefundQuotaInvest(db, invest)
	}
	return nil, nil
}

func handleDeFiReceiveErr(logger log15.Logger, method string, err error, sendBlock *ledger.AccountBlock) ([]*ledger.AccountBlock, error) {
	logger.Error("deFi receive with err", "error", err.Error(), "method", method, "sendBlockHash", sendBlock.Hash.String(), "sendAddress", sendBlock.AccountAddress.String())
	return nil, err
}